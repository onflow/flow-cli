pub contract HelloWorld {

    pub event FooEvent(x: [String]?)
    pub let greeting: String
    pub var arrayEvent: [String]
    init() {
        self.greeting = "Helloo, Worlds!"
        self.arrayEvent = ["Event"]
    }

    pub fun hello(): String {
        emit FooEvent(x: self.arrayEvent)
        return self.greeting
    }
}