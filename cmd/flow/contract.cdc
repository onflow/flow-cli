pub contract HelloWorld {

    pub event FooEvent(x: [String])
    //pub event FooFooFooFooEvent(x: String)

    pub let greeting: String
    pub var arrayEvent: [String]
    init() {
        self.greeting = "Hello, World!"
        self.arrayEvent = ["Event"]
    }

    pub fun hello(): String {
        emit FooEvent(x: self.arrayEvent)
        //emit FooFooFooFooEvent(x: self.greeting)
        return self.greeting
    }
}