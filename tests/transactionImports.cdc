import Bar from "./Bar.cdc"

transaction() {
  let guest: Address

  prepare(authorizer: AuthAccount) {
    self.guest = authorizer.address
  }

  execute {
    log(self.guest.toString())
  }
}