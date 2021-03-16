transaction() {
  prepare(authorizer: AuthAccount) {}

  execute {
    panic("Error error")
  }
}