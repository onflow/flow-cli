pub contract HelloWorldNew {

    pub event FooEvent(x: [String]?)
    pub let greeting: String
    pub var arrayEvent: [String]
    init() {
        self.greeting = "Hello, World New!"
        self.arrayEvent = ["Event"]
    }

    pub fun hello(): String {
        emit FooEvent(x: self.arrayEvent)
        return self.greeting
    }
}