import HelloWorld from 0xf8d6e0586b0a20c7

transaction {

    // No need to do anything in prepare because we are not working with
    // account storage.
	prepare(acct: AuthAccount) {}

    // In execute, we simply call the hello function
    // of the HelloWorld contract and log the returned String.
	execute {
	  	log(HelloWorld.hello())
	}
}