import Foo from "./Foo.cdc"

pub contract Bar {
    init(a: String, b: UInt32) {
        log(a.concat(b.toString()))
    }
}
