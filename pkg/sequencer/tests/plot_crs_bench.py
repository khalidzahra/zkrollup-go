import csv
import matplotlib.pyplot as plt
import sys

# Usage: python plot_crs_bench.py [csv_file] [output_image]
# Defaults: csv_file='crs_bench_results.csv', output_image='crs_bench_results.png'

def main():
    csv_file = sys.argv[1] if len(sys.argv) > 1 else 'crs_bench_results.csv'
    output_image = sys.argv[2] if len(sys.argv) > 2 else 'crs_bench_results.png'

    participants = []
    durations = []

    with open(csv_file, newline='') as f:
        reader = csv.DictReader(f)
        for row in reader:
            participants.append(int(row['participants']))
            durations.append(float(row['duration_seconds']))

    plt.figure(figsize=(8,5))
    plt.plot(participants, durations, marker='o', linestyle='-')
    plt.xlabel('Number of Participants')
    plt.ylabel('Ceremony Duration (seconds)')
    plt.title('CRS Ceremony Performance')
    plt.grid(True)
    plt.tight_layout()
    plt.savefig(output_image)
    print(f"Graph saved to {output_image}")

if __name__ == '__main__':
    main()
