
!!!THIS README WAS GENERATED USING CHATGPT!!!

Distributed System with gRPC
This is a simple example of a distributed system using gRPC for communication between nodes. The system allows nodes to request access to a critical section and demonstrates a basic form of distributed coordination using Lamport timestamps.

Prerequisites
Go installed on your machine.
Ensure that your Go environment is properly set up.
Running the Program
Clone the repository:

bash

git clone https://github.com/your-username/your-repository.git
cd your-repository
Generate gRPC code:

bash

protoc grpc/proto.proto --go_out=plugins=grpc:grpc
Build the executable:

bash

go build -o node main.go
Run the nodes:

Open three separate terminals and run the following commands to start three nodes:

bash

./node 1 50051 127.0.0.1:50052 127.0.0.1:50053
bash

./node 2 50052 127.0.0.1:50051 127.0.0.1:50053
bash

./node 3 50053 127.0.0.1:50051 127.0.0.1:50052
Replace 127.0.0.1 with the actual IP addresses or hostnames if running on different machines.

Observe the system logs:

The nodes will exchange messages, and you will see logs indicating when a node enters and leaves the critical section.

Notes
Ensure that there are no firewall issues preventing communication between nodes.
Feel free to experiment with different scenarios and configurations.
Have fun experimenting with your distributed system!


!!!THIS README WAS GENERATED USING CHATGPT!!!