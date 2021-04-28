import Foo from "./Foo.cdc"

pub contract Bar {
    init(a: String, b: String) {
        log(a.concat(b))
    }
}
