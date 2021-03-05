package cmd

type Result interface {
	String() string
	Oneliner() string
	JSON() interface{}
}
