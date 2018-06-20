package p2p

import (
	"context"

	"gx/ipfs/QmNqRnejxJxjRroz7buhrjfU8i3yNBLa81hFtmf2pXEffN/go-multiaddr-net"
	ma "gx/ipfs/QmUxSEGbv2nmYNnfXi7839wwQqTN3kwQeUxe8dTjZWZs7J/go-multiaddr"
	"gx/ipfs/QmXdgNhVEgjLxjUoMs5ViQL7pboAt3Y7V7eGHRiE4qrmTE/go-libp2p-net"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
)

// remoteListener accepts libp2p streams and proxies them to a manet host
type remoteListener struct {
	p2p *P2P

	// Application proto identifier.
	proto protocol.ID

	// Address to proxy the incoming connections to
	addr ma.Multiaddr
}

// ForwardRemote creates new p2p listener
func (p2p *P2P) ForwardRemote(ctx context.Context, proto protocol.ID, addr ma.Multiaddr) (Listener, error) {
	listener := &remoteListener{
		p2p: p2p,

		proto: proto,
		addr:  addr,
	}

	if err := p2p.Listeners.Register(listener); err != nil {
		return nil, err
	}

	return listener, nil
}

func (l *remoteListener) start() error {
	// TODO: handle errors when https://github.com/libp2p/go-libp2p-host/issues/16 will be done
	l.p2p.peerHost.SetStreamHandler(l.proto, func(remote net.Stream) {
		local, err := manet.Dial(l.addr)
		if err != nil {
			remote.Reset()
			return
		}

		peerMa, err := ma.NewMultiaddr("/ipfs/" + remote.Conn().RemotePeer().Pretty())
		if err != nil {
			remote.Reset()
			return
		}

		stream := &Stream{
			Protocol: l.proto,

			OriginAddr: peerMa,
			TargetAddr: l.addr,

			Local:  local,
			Remote: remote,

			Registry: l.p2p.Streams,
		}

		l.p2p.Streams.Register(stream)
		stream.startStreaming()
	})

	return nil
}

func (l *remoteListener) Protocol() protocol.ID {
	return l.proto
}

func (l *remoteListener) ListenAddress() ma.Multiaddr {
	addr, err := ma.NewMultiaddr("/ipfs/" + l.p2p.identity.Pretty())
	if err != nil {
		panic(err)
	}
	return addr
}

func (l *remoteListener) TargetAddress() ma.Multiaddr {
	return l.addr
}

func (l *remoteListener) Close() error {
	l.p2p.peerHost.RemoveStreamHandler(l.proto)
	l.p2p.Listeners.Deregister(getListenerKey(l))
	return nil
}