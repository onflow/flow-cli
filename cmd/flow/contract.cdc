pub contract HelloWorld {

    pub event FooEvent(x: [String]?)
    pub let greeting: String
    pub var arrayEvent: [String]
    pub let id: UInt64
    init(newID: UInt64) {
        self.greeting = "Hello, World!"
        self.id = newID
        self.arrayEvent = ["Event"]
    }

    pub fun hello(): String {
        emit FooEvent(x: self.arrayEvent)
        return self.greeting
    }
}