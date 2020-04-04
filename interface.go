package hdwallet

type IExtendKey interface {
	Child(index uint32) (IExtendKey, error)
	String() string
	ExtendKeyStr() string
	Address(param interface{}) (Address, error)
	IsPrivate() bool
	Base() *BaseExtKey
}
