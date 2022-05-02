package user

type RequiredUserFields struct {
	FirstName string `bson:"firstName"`
	LastName  string `bson:"lastName"`
	UserName  string `bson:"userName"`
	Email     string `bson:"email"`
	Region    string `bson:"region"`
	Password  string `bson:"password"`
}