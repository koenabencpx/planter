
# Beforehand 
 
Download latest plantuml release from github and save in root of project as plantuml.jar

## Quick Start on windows

`go build`

## 1.A. export entire database into PlantUml-txt 
`./planter.exe postgres://postgres:postgres@localhost/wvp?sslmode=disable -o database_name.uml.txt` 

## 1.B. export module with a specified set of tables into PlantUml-txt
`./planter.exe postgres://postgres:postgres@localhost/wvp?sslmode=disable -o module.uml.txt -m ./module_programmering.tables.txt`


## 2. finally, generate png image from PlantUml via this command:   
`java -jar plantuml.jar -verbose module.uml.txt`

