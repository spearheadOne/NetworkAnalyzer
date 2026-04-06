#!/usr/bin/env python3

from mininet.topo import Topo
from mininet.net import Mininet
from mininet.node import OVSSwitch
from mininet.cli import CLI
from mininet.node import OVSBridge

class BaseTopo(Topo):
    def build(self):
        h1 = self.addHost("host1", ip="10.0.0.1/24")
        h2 = self.addHost("host2", ip="10.0.0.2/24")

        s1 = self.addSwitch("switch1")
        s2 = self.addSwitch("switch2")
        s3 = self.addSwitch("switch3")

        self.addLink(h1,s1)
        self.addLink(s1,s2)
        self.addLink(s2,s3)
        self.addLink(s3,h2)



if __name__ == "__main__":
    topo = BaseTopo()
    net = Mininet(topo=topo, switch=OVSBridge, controller=None)

    net.start()
    CLI(net)
    net.stop()