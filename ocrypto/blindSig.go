package ocrypto

type BlindSig struct {

}

func (bs *BlindSig) MaskSingle(*[]byte) {}
func (bs *BlindSig) MaskMultiple(*[][]byte) {}
func (bs *BlindSig) UnmaskSingle(*[]byte) {}
func (bs *BlindSig) UnmaskMultiple(*[][]byte) {}