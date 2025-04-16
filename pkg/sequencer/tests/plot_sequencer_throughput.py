import csv
import matplotlib.pyplot as plt

# Read CSV data
tx_counts = []
durations = []
thruputs = []
with open('sequencer_throughput.csv', newline='') as csvfile:
    reader = csv.DictReader(csvfile)
    for row in reader:
        tx_counts.append(int(row['transactions']))
        durations.append(float(row['duration_seconds']))
        thruputs.append(float(row['throughput_tps']))

# Plot throughput vs transactions
plt.figure(figsize=(8, 5))
plt.plot(tx_counts, thruputs, marker='o', linestyle='-', color='b')
plt.title('Sequencer Throughput Benchmark')
plt.xlabel('Number of Transactions')
plt.ylabel('Throughput (tx/sec)')
plt.grid(True)
plt.tight_layout()
plt.savefig('sequencer_throughput.png')
plt.show()
