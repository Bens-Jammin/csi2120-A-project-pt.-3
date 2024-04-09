# Ben Miller - 300297574
# Arin Barak - 300280812
# This file is just to display the experiment data

import pandas as pd
from matplotlib import pyplot as plt

# data from tests
data = {
    '1': [28106, 24560, 24227, 25721, 24280, 23090, 23160, 24386, 24976, 24347],
    '2': [15648, 14758, 13591, 13080, 13223, 17428, 15346, 14080, 13613, 17812],
    '4': [8856, 8694, 8581, 8670, 8470, 9427, 8513, 8674, 8787, 9601],
    '16': [6983, 6483, 6414, 6417, 6443, 6711, 6343, 6368, 6606, 6518],
    '64': [6452, 6727, 6559, 6326, 6526, 6441, 6575, 6768, 6809, 6459],
    '256': [7646, 7094, 7299, 6886, 6965, 7239, 6988, 7290, 7190, 8089],
    '1048': [8849, 8779, 8136, 8628, 8635, 8780, 9228, 9142, 8671, 8524]
}

df = pd.DataFrame(data)
plt.figure(figsize=(10, 6))
df.boxplot(column=['1', '2', '4', '16', '64', '256', '1048'])
plt.title('Image Relatedness Runtime With Different Number of Goroutines')
plt.xlabel('Number of Goroutines')
plt.ylabel('Time (ms)')
plt.xticks(rotation=45)
plt.grid(True)
plt.tight_layout()
plt.savefig("experiment_results.png")
plt.show()