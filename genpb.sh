#!/bin/bash

protoc phonebookpb/phonebook.proto --go_out=plugins=grpc:.