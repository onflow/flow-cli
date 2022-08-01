pub contract HelloWorld {

    pub event FooEvent(x: [String]?)
    pub let greeting: String
    pub var arrayEvent: [String]
    init() {
        self.greeting = "Hello, World!"
        self.arrayEvent = ["Event"]
    }

    pub fun hello(word: String): String {
        emit FooEvent(x: self.arrayEvent)
        self.greeting = word
        return self.greeting
    }
}