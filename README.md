# NTHU-Distributed-System

The repository includes modules for the NTHU Distributed System course lab with mono-repo architecture.

## Code Generation

Some modules use gRPC for communication or use the `mockgen` library for unit testing.

So there is a need to generate code manually when the code changed.

Before generating code, make sure your Docker is running since we are generating code inside a Docker container to prevent dependencies from conflicting/missing on your host machine.

For generating code for all modules, run `make generate`.

For generating code for a single module, run `make {module}.generate`. For example: `make video.generate`.
