# HashServer_JC
JC Assignment

How to run HashServer

To run this application, open the cmd prompt and execute command 
go run HashService.go

Alternatively you can add a server:port as parameter and execute command as

go run HashService.go :8080 

or 

go run HashService.go localhost:8080

By default, the application will use localhost:8080 if no value is supplied.


Other considerations:
This application should have ben into multiple classes/modules

The first class would be the HashServer itself which would be responsible for handling traffic and the second class would be a HashManager. The has manager would act as a database or datastore and would allow the design to abstract out the implementation of the HashManager. The HashManager would be called by they HashServer to handle the task of hashing the password, returning the hashed password, and reporting its stats. 

This was a fun experience and thank you for the opportunity.

