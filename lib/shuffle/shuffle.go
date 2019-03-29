package libunlynxshuffle

import (
	"crypto/cipher"
	"math/big"
	"os"
	"sync"

	"github.com/lca1/unlynx/lib"
	"github.com/lca1/unlynx/lib/tools"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/util/random"
	"go.dedis.ch/onet/v3/log"
)

// ShuffleSequence applies shuffling to a ciphervector
func ShuffleSequence(inputList []libunlynx.CipherVector, g, h kyber.Point, precomputed []CipherVectorScalar) ([]libunlynx.CipherVector, []int, [][]kyber.Scalar) {
	maxUint := ^uint(0)
	maxInt := int(maxUint >> 1)

	// number of elgamal pairs
	NQ := len(inputList[0])
	k := len(inputList) // number of clients

	rand := libunlynx.SuiTe.RandomStream()
	// Pick a fresh (or precomputed) ElGamal blinding factor for each pair
	beta := make([][]kyber.Scalar, k)
	precomputedPoints := make([]libunlynx.CipherVector, k)
	for i := 0; i < k; i++ {
		if precomputed == nil {
			beta[i] = libunlynx.RandomScalarSlice(NQ)
		} else {
			randInt := random.Int(big.NewInt(int64(maxInt)), rand)

			indice := int(randInt.Int64() % int64(len(precomputed)))
			beta[i] = precomputed[indice].S[0:NQ] //if beta file is bigger than query line responses
			precomputedPoints[i] = precomputed[indice].CipherV[0:NQ]
		}

	}

	// Pick a random permutation
	pi := libunlynx.RandomPermutation(k)

	outputList := make([]libunlynx.CipherVector, k)

	wg := libunlynx.StartParallelize(k)
	for i := 0; i < k; i++ {
		go func(outputList []libunlynx.CipherVector, i int) {
			defer wg.Done()
			shuffle(pi, i, inputList, outputList, NQ, beta, precomputedPoints, g, h)
		}(outputList, i)
	}
	libunlynx.EndParallelize(wg)

	return outputList, pi, beta
}

// shuffle applies shuffling and rerandomization
func shuffle(pi []int, i int, inputList, outputList []libunlynx.CipherVector, NQ int, beta [][]kyber.Scalar, precomputedPoints []libunlynx.CipherVector, g, h kyber.Point) {
	index := pi[i]
	outputList[i] = *libunlynx.NewCipherVector(NQ)
	wg := libunlynx.StartParallelize(NQ)
	for j := 0; j < NQ; j++ {
		var b kyber.Scalar
		var cipher libunlynx.CipherText
		if len(precomputedPoints[0]) == 0 {
			b = beta[index][j]
		} else {
			cipher = precomputedPoints[index][j]
		}
		go func(j int) {
			defer wg.Done()
			outputList[i][j] = rerandomize(inputList[index], b, b, cipher, g, h, j)
		}(j)
	}
	libunlynx.EndParallelize(wg)
}

// rerandomize rerandomizes an element in a ciphervector at position j, following the Neff Shuffling algorithm
func rerandomize(cv libunlynx.CipherVector, a, b kyber.Scalar, cipher libunlynx.CipherText, g, h kyber.Point, j int) libunlynx.CipherText {
	ct := libunlynx.NewCipherText()
	var tmp1, tmp2 kyber.Point

	if cipher.C == nil {
		//no precomputed value
		tmp1 = libunlynx.SuiTe.Point().Mul(a, g)
		tmp2 = libunlynx.SuiTe.Point().Mul(b, h)
	} else {
		tmp1 = cipher.K
		tmp2 = cipher.C
	}

	ct.K = libunlynx.SuiTe.Point().Add(cv[j].K, tmp1)
	ct.C = libunlynx.SuiTe.Point().Add(cv[j].C, tmp2)
	return *ct
}

// Precomputation
//______________________________________________________________________________________________________________________

// CreatePrecomputedRandomize creates precomputed values for shuffling using public key and size parameters
func CreatePrecomputedRandomize(g, h kyber.Point, rand cipher.Stream, lineSize, nbrLines int) []CipherVectorScalar {
	result := make([]CipherVectorScalar, nbrLines)
	wg := libunlynx.StartParallelize(len(result))
	var mutex sync.Mutex
	for i := range result {
		result[i].CipherV = make(libunlynx.CipherVector, lineSize)
		result[i].S = make([]kyber.Scalar, lineSize)

		go func(i int) {
			defer (*wg).Done()

			for w := range result[i].CipherV {
				mutex.Lock()
				tmp := libunlynx.SuiTe.Scalar().Pick(rand)
				mutex.Unlock()

				result[i].S[w] = tmp
				result[i].CipherV[w].K = libunlynx.SuiTe.Point().Mul(tmp, g)
				result[i].CipherV[w].C = libunlynx.SuiTe.Point().Mul(tmp, h)
			}

		}(i)
	}
	libunlynx.EndParallelize(wg)
	return result
}

// PrecomputeForShuffling precomputes data to be used in the shuffling protocol (to make it faster) and saves it in a .gob file
func PrecomputeForShuffling(serverName, gobFile string, surveySecret kyber.Scalar, collectiveKey kyber.Point, lineSize int) []CipherVectorScalar {
	log.Lvl1(serverName, " precomputes for shuffling")
	scalarBytes, _ := surveySecret.MarshalBinary()
	precomputeShuffle := CreatePrecomputedRandomize(libunlynx.SuiTe.Point().Base(), collectiveKey, libunlynx.SuiTe.XOF(scalarBytes), lineSize*2, 10)

	encoded, err := EncodeCipherVectorScalar(precomputeShuffle)

	if err != nil {
		log.Error("Error during marshaling")
	}
	libunlynxtools.WriteToGobFile(gobFile, encoded)

	return precomputeShuffle
}

// PrecomputationWritingForShuffling reads the precomputation data from  .gob file if it already exists or generates a new one
func PrecomputationWritingForShuffling(appFlag bool, gobFile, serverName string, surveySecret kyber.Scalar, collectiveKey kyber.Point, lineSize int) []CipherVectorScalar {
	log.Lvl1(serverName, " precomputes for shuffling")
	var precomputeShuffle []CipherVectorScalar
	if appFlag {
		if _, err := os.Stat(gobFile); os.IsNotExist(err) {
			precomputeShuffle = PrecomputeForShuffling(serverName, gobFile, surveySecret, collectiveKey, lineSize)
		} else {
			var encoded []CipherVectorScalarBytes
			libunlynxtools.ReadFromGobFile(gobFile, &encoded)

			precomputeShuffle, err = DecodeCipherVectorScalar(encoded)

			if len(precomputeShuffle[0].CipherV) < lineSize {

			}
			if err != nil {
				log.Error("Error during unmarshaling")
			}
		}
	} else {
		scalarBytes, _ := surveySecret.MarshalBinary()
		precomputeShuffle = CreatePrecomputedRandomize(libunlynx.SuiTe.Point().Base(), collectiveKey, libunlynx.SuiTe.XOF(scalarBytes), lineSize*2, 10)
	}
	return precomputeShuffle
}

// ReadPrecomputedFile reads the precomputation data from a .gob file
func ReadPrecomputedFile(fileName string) []CipherVectorScalar {
	var precomputeShuffle []CipherVectorScalar
	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		var encoded []CipherVectorScalarBytes
		libunlynxtools.ReadFromGobFile(fileName, &encoded)

		precomputeShuffle, _ = DecodeCipherVectorScalar(encoded)
	} else {
		precomputeShuffle = nil
	}
	return precomputeShuffle
}
