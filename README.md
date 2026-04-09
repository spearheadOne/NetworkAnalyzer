# Network Analyzer

## TODO - RUN
sudo python3 topology.py
sudo ./topology.py

host1 iperf -c 10.0.0.2 -t 20 -P 5
host2 iperf -c 10.0.0.1 -t 20 -P 5
host2 iperf -s &

sudo mn -c
sudo tcpdump -ni any udp port 6343