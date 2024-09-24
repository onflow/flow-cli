import "Counter"

transaction {
    prepare(signer: auth(BorrowValue) &Account) {
        // Borrow a reference to the <no value> contract's public capability
        let counterRef = signer.storage.borrow<&Counter>(from: /storage/counter)
            ?? panic("Could not borrow reference to the counter")

        // Call the increment function on the Counter contract
        counterRef.increment()
    }

    execute {
        log("Counter incremented successfully")
    }
}