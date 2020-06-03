package random

// Rand is a pseudo random number generator
type Rand interface {
	// IntN returns a int random number between 0 and "to" (exclusive)
	IntN(int) (int, error)

	// Permutation returns a permutation of the set [0,n-1]
	// the theoretical output space grows very fast with (!n) so that input (n) should be chosen carefully
	// to make sure the function output space covers a big chunk of the theoretical outputs.
	// The returned error is non-nil if the parameter is a negative integer.
	Permutation(n int) ([]int, error)

	// SubPermutation returns the m first elements of a permutation of [0,n-1]
	// the theoretical output space can be large (n!/(n-m)!) so that the inputs should be chosen carefully
	// to make sure the function output space covers a big chunk of the theoretical outputs.
	// The returned error is non-nil if the parameter is a negative integer.
	SubPermutation(n int, m int) ([]int, error)

	// Shuffle permutes an ordered data structure of an arbitrary type in place. The main use-case is
	// permuting slice or array elements. (n) is the size of the data structure.
	// the theoretical output space grows very fast with the slice size (n!) so that input (n) should be chosen carefully
	// to make sure the function output space covers a big chunk of the theoretical outputs.
	// The returned error is non-nil if any of the parameters is a negative integer.
	Shuffle(n int, swap func(i, j int)) error

	// Samples picks (m) random ordered elements of a data structure of an arbitrary type of total size (n). The (m) elements are placed
	// in the indices 0 to (m-1) with in place swapping. The data structure ends up being a permutation of the initial (n) elements.
	// While the sampling of the (m) elements is pseudo-uniformly random, there is no guarantee about the uniformity of the permutation of
	// the (n) elements. The function Shuffle should be used in case the entire (n) elements need to be shuffled.
	// The main use-case of the data structure is a slice or array.
	// The theoretical output space grows very fast with the slice size (n!/(n-m)!) so that inputs should be chosen carefully
	// to make sure the function output space covers a big chunk of the theoretical outputs.
	// The returned error is non-nil if any of the parameters is a negative integer.
	Samples(n int, m int, swap func(i, j int)) error

	// State returns the internal state of the random generator.
	// The internal state can be used as a seed input for the function
	// NewRand to restore an identical PRG (with the same internal state)
	State() []byte
}
