# geeORM

a simple ORM based on go-mysql-driver, It references gORM and realize it's main function.

Until now, it is only support mysql database.

It every function has it own test.go file, you need to set your mysql address in function OpenDB in file geeorm_test.go.



The project use reflect package to realize Insert , CreatTable and hook function.

It is also support transaction, you can practice by TestEngine_Transaction function in geeorm_test.go.