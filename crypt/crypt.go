package crypt

//face info
type Crypt struct {
	simple *SimpleEncrypt
	jwt *Jwt
}

//construct
func NewCrypt() *Crypt {
	this := &Crypt{
		simple: NewSimpleEncrypt(),
		jwt: NewJwt(),
	}
	return this
}

//get sub instance
func (f *Crypt) GetMongo() *Jwt {
	return f.jwt
}

func (f *Crypt) GetSimple() *SimpleEncrypt {
	return f.simple
}

