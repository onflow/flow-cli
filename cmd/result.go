package cmd

type Result interface {
	String() string
	JSON() string
}
