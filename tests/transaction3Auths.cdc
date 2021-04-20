transaction(greeting: String) {
  let guest: Address

  prepare(authorizer1: AuthAccount, authorizer2: AuthAccount, authorizer3: AuthAccount) {
    self.guest = authorizer1.address
  }

  execute {
    log(greeting.concat(",").concat(self.guest.toString()))
  }
}