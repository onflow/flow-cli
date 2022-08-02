pub contract HelloWorld {

    pub event FooEvent(x: String)
    pub var greeting: String
    init(a:String) {
    //init() {
        //self.greeting = "Hello World"
        self.greeting = a
    }

    pub fun hello(): String {
        return self.greeting
    }
}