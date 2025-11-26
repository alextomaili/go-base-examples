Как парсить TCP пакеты
github.com/google/gopacket

func (r *Reader) Process() error {
	p := gopacket.NewPacket(r.buf[:r.n], layers.LayerTypeIPv4, gopacket.NoCopy)
	tcpLayer := p.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return errors.New("packet doesn't contain LayerTypeTCP")
	}

	tcp := tcpLayer.(*layers.TCP)
	r.strDumpPacket(tcp)
	return nil
}


Как еще можно парсить пакеты
https://github.com/mikioh/tcpinfo

