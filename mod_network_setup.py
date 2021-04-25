from mininet.topo import Topo
from mininet.net import Mininet
from mininet.node import Node
from mininet.log import setLogLevel, info
from mininet.cli import CLI


class LinuxRouter( Node ):
    "A Node with IP forwarding enabled."

    # pylint: disable=arguments-differ
    def config( self, **params ):
        super( LinuxRouter, self).config( **params )
        # Enable forwarding on the router
        self.cmd( 'sysctl net.ipv4.ip_forward=1' )

    def terminate( self ):
        self.cmd( 'sysctl net.ipv4.ip_forward=0' )
        super( LinuxRouter, self ).terminate()

class MPTopo(Topo):
    def build(self):
        IP1 = '192.168.1.1/24'
        host = [['192.168.1.2/24', '192.168.1.100/24'], 
            ['192.168.1.3/24', '192.168.1.101/24']]
        router = self.addNode( 'r0', cls=LinuxRouter, ip=IP1 )
        s1, s2 = [ self.addSwitch( s ) for s in ( 's1', 's2' ) ]
        self.addLink( s1, router, intfName2='r0-eth1',
                      params2={ 'ip' : IP1 } )
        self.addLink( s2, router, intfName2='r0-eth2',
                      params2={ 'ip' : IP1 } )
        h1 = self.addHost('h1',ip=host[0][0], defaultRoute='via {}'.format(IP1))
        h2 = self.addHost('h2',ip=host[1][0], defaultRoute='via {}'.format(IP1))
        for i, h in enumerate([h1, h2]):
            for j, s in enumerate([s1, s2]):
                self.addLink(h,s, params1={'ip':host[i][j]})

def run():
    net = Mininet(topo=MPTopo(), waitConnected=True)
    net.start()
    info( '*** Routing Table on Router:\n' )
    info( net[ 'r0' ].cmd( 'route' ) )
    CLI( net )
    net.stop()

if __name__ == '__main__':
    setLogLevel('info')
    run()
