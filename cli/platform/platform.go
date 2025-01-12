package platform

type Client interface {
	Logg(bool, bool, bool, int)
	ProxyOn()
	ProxyOff()
}
