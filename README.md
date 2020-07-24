# Boatswain

Boatswain is a PoC implementation for doing EKS node maintenance/upgrades by
replacing nodes which were created from outdated launch templates.

## Docker images

Docker images are built automatically from tags and the master branch and
pushed to
[DockerHub](https://hub.docker.com/repository/docker/projectsyn/boatswain).

The `latest` tag is built from the master branch, and tagged images are built
from Git tags.

## Documentation

Documentation for this component is written using [Asciidoc][asciidoc] and [Antora][antora].
It is located in the [docs/](docs) folder.
The [Divio documentation structure](https://documentation.divio.com/) is used to organize its content.

## Contributing and license

This library is licensed under [BSD-3-Clause](LICENSE).
For information about how to contribute see [CONTRIBUTING](CONTRIBUTING.md).

[asciidoc]: https://asciidoctor.org/
[antora]: https://antora.org/
