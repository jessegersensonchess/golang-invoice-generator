# Invoice Generator
Invoice generator is a command line based tool that reads data from CSV and generates an invoice using package gofpdf https://github.com/jung-kurt/gofpdf

![image](https://user-images.githubusercontent.com/49806519/119263878-6af87380-bc13-11eb-99ab-3bbf0d085c34.png)

## Quick Start

```
git clone https://github.com/jessegersensonchess/golang-invoice-generator
cd golang-invoice-generator
#### edit constants in invoice.go
docker build -t invoice:latest 
docker run -it --rm -v/tmp/invoice:/tmp invoice:latest invoice -k 20.99 generate /tmp/test.csv
```

Output invoice will be in /tmp/invoice.   

Navigate to the directory and run the following command to generate an invoice based on the sample input data
``` shell
./invoice-generator generate ./sample/invoice.csv -p U-1423242 -n INV-42532622
```

For more modification on the billing information, run
``` shell
./invoice-generator generate -h  
```
