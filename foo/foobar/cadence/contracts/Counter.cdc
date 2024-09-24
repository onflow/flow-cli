access(all) contract Counter {

    access(all) var count: Int

    // Event to be emitted when the counter is incremented
    access(all) event CounterIncremented(newCount: Int)

    // Event to be emitted when the counter is decremented
    access(all) event CounterDecremented(newCount: Int)

    init() {
        self.count = 0
    }

    // Public function to increment the counter
    access(all) fun increment() {
        self.count = self.count + 1
        emit CounterIncremented(newCount: self.count)
    }

    // Public function to decrement the counter
    access(all) fun decrement() {
        self.count = self.count - 1
        emit CounterDecremented(newCount: self.count)
    }

    // Public function to get the current count
    view access(all) fun getCount(): Int {
        return self.count
    }
}