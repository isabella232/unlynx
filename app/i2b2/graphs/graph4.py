import matplotlib.pyplot as plt
import numpy as np
import pandas as pd

raw_data_query_one = {
    'x_label':  ['3', '6', '9', '10'],
    'y1_label': [2, 2, 2, 2],                   # Insecure i2b2
    'y2_label': [2.4, 2.7, 3, 3.1],             # Without DDT
    'y3_label': [90.9, 287.2, 544.3, 606.3],    # With DDT
}

font = {'family': 'Bitstream Vera Sans',
        'size': 22}

plt.rc('font', **font)

df = pd.DataFrame(raw_data_query_one, raw_data_query_one['x_label'])

N = 4
ind = np.arange(N)  # The x locations for the groups

# Create the general blog and the "subplots" i.e. the bars
fig, ax1 = plt.subplots(1, figsize=(14, 12))

# Set the bar width
bar_width = 0.24

# Container of all bars
bars = []

# Create a bar plot, in position bar_l
bars.append(ax1.bar(ind,
                    # using the y1_label data
                    df['y1_label'],
                    # set the width
                    width=bar_width,
                    label='Insecure i2b2',
                    # with alpha 1
                    alpha=0.5,
                    # with color
                    color='#1abc78'))

# Create a bar plot, in position bar_l
bars.append(ax1.bar(ind + bar_width,
                    # using the y3_label data
                    df['y2_label'],
                    # set the width
                    width=bar_width,
                    label='MedCo',
                    # with alpha 1
                    alpha=0.5,
                    # with color
                    color='#3232FF'))

# Create a bar plot, in position bar_l
bars.append(ax1.bar(ind + bar_width + bar_width,
                    # using the y2_label data
                    df['y3_label'],
                    # set the width
                    width=bar_width,
                    label='MedCo+',
                    # with alpha 1
                    alpha=0.5,
                    # with color
                    color='#bc1a1a'))

# Set the x ticks with names
ax1.set_xticks(ind + bar_width + bar_width/2)
ax1.set_xticklabels(df['x_label'])
ax1.set_yscale('symlog', basey=10)
ax1.yaxis.grid(True)
ax1.set_ylim([0, 100000])
ax1.set_xlim([0, ind[3] + bar_width + bar_width + bar_width])

# Labelling
ax1.text(ind[0] + bar_width/2 - 0.03, df['y1_label'][0]+0.2,
         str(df['y1_label'][0]), color='black', fontweight='bold')
ax1.text(ind[1] + bar_width/2 - 0.03, df['y1_label'][1]+0.2,
         str(df['y1_label'][1]), color='black', fontweight='bold')
ax1.text(ind[2] + bar_width/2 - 0.03, df['y1_label'][2]+0.2,
         str(df['y1_label'][2]), color='black', fontweight='bold')
ax1.text(ind[3] + bar_width/2 - 0.03, df['y1_label'][3]+0.2,
         str(df['y1_label'][3]), color='black', fontweight='bold')

ax1.text(ind[0] + bar_width + bar_width/2 - 0.09, df['y2_label'][0]+0.2,
         str(df['y2_label'][0]), color='black', fontweight='bold')
ax1.text(ind[1] + bar_width + bar_width/2 - 0.09, df['y2_label'][1]+0.2,
         str(df['y2_label'][1]), color='black', fontweight='bold')
ax1.text(ind[2] + bar_width + bar_width/2 - 0.09, df['y2_label'][2]+0.2,
         str(df['y2_label'][2]), color='black', fontweight='bold')
ax1.text(ind[3] + bar_width + bar_width/2 - 0.09, df['y2_label'][3]+0.2,
         str(df['y2_label'][3]), color='black', fontweight='bold')

ax1.text(ind[0] + bar_width + bar_width + bar_width/2 - 0.13, df['y3_label'][0] + df['y3_label'][0]/10,
         str(df['y3_label'][0]), color='black', fontweight='bold')
ax1.text(ind[1] + bar_width + bar_width + bar_width/2 - 0.17, df['y3_label'][1] + df['y3_label'][1]/10,
         str(df['y3_label'][1]), color='black', fontweight='bold')
ax1.text(ind[2] + bar_width + bar_width + bar_width/2 - 0.17, df['y3_label'][2] + df['y3_label'][2]/10,
         str(df['y3_label'][2]), color='black', fontweight='bold')
ax1.text(ind[3] + bar_width + bar_width + bar_width/2 - 0.22, df['y3_label'][3] + df['y3_label'][3]/10,
         str(df['y3_label'][3]), color='black', fontweight='bold')

# Set the label and legends
ax1.set_ylabel("Runtime (s)", fontsize=22)
ax1.set_xlabel("Total number of servers", fontsize=22)
plt.legend(loc='upper left')

ax1.tick_params(axis='x', labelsize=22)
ax1.tick_params(axis='y', labelsize=22)

plt.savefig('scalability_#servers.pdf', format='pdf')
