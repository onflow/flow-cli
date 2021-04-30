pub contract Hello {
    pub let greeting: String
    init() {
        self.greeting = "Hello, World V2!"
    }
    pub fun hello(): String {
        return self.greeting
    }
}
