package crypt

//face info
type Crypt struct {
	simple *SimpleEncrypt
	jwt *Jwt
	rsa *Rsa
}

//construct
func NewCrypt() *Crypt {
	this := &Crypt{
		simple: NewSimpleEncrypt(),
		jwt: NewJwt(),
		rsa: NewRsa(),
	}
	return this
}

//get sub instance
func (f *Crypt) GetRsa() *Rsa {
	return f.rsa
}

func (f *Crypt) GetJwt() *Jwt {
	return f.jwt
}

func (f *Crypt) GetSimple() *SimpleEncrypt {
	return f.simple
}

