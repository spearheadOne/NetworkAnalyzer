#!/usr/bin/env python3

from mininet.cli import CLI
from mininet.net import Mininet
from mininet.node import OVSBridge
from mininet.node import OVSSwitch
from mininet.topo import Topo

COLLECTOR_HOST = "host.lima.internal"  # replace if needed
COLLECTOR_PORT = 6343


class BaseTopo(Topo):
    def build(self):
        h1 = self.addHost("host1", ip="10.0.0.1/24")
        h2 = self.addHost("host2", ip="10.0.0.2/24")

        s1 = self.addSwitch("switch1")
        s2 = self.addSwitch("switch2")
        s3 = self.addSwitch("switch3")

        self.addLink(h1, s1)
        self.addLink(s1, s2)
        self.addLink(s2, s3)
        self.addLink(s3, h2)


def enable_sflow(net, collector_host, collector_port, sampling=10, polling=20):
    for sw in net.switches:
        bridge = sw.name
        agent_if = f"{bridge}-eth1"

        cmd = (
            f'ovs-vsctl -- --id=@sflow create sflow '
            f'agent={agent_if} '
            f'targets=\\"{collector_host}:{collector_port}\\" '
            f'sampling={sampling} polling={polling} '
            f'-- set bridge {bridge} sflow=@sflow'
        )

        print(f"\nConfiguring sFlow on {bridge}")
        print(cmd)
        out = sw.cmd(cmd)
        if out.strip:
            print(out)


if __name__ == "__main__":
    topo = BaseTopo()
    net = Mininet(topo=topo, switch=OVSBridge, controller=None)

    net.start()

    enable_sflow(net, collector_host=COLLECTOR_HOST, collector_port=COLLECTOR_PORT)

    CLI(net)
    net.stop()
