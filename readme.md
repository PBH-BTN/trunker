# BTN-Server

## Introduce

A demo with `Hertz` and `Gorm`

- Use `thrift` IDL to define `HTTP` interface
- Use `hz` to generate code
- Use `Hertz` binding and validate
- Use `Gorm` and `MySQL`

## IDL

This project use `thrift` IDL to define `HTTP` interface. The specific interface define in [api.thrift](idl/api.thrift)

## Code generation tool

This project use `hz` to generate code. The use of `hz` refers
to [hz](https://www.cloudwego.io/docs/hertz/tutorials/toolkit/toolkit/)

The `hz` commands used can be found in [Makefile](Makefile)

## Binding and Validate

The use of binding and Validate refers
to [Binding and Validate](https://www.cloudwego.io/docs/hertz/tutorials/basic-feature/binding-and-validate/)

## Gorm

This project use `Gorm` to operate `MySQL` and refers to [Gorm](https://gorm.io/)

## How to run

### Run mysql docker

```bash
docker-compose up
```

### Run

```bash
./build.sh
cd output
./bootstrap.sh
```