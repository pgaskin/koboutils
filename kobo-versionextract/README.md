# kobo-versionextract
Extracts the version from an update zip or a KoboRoot.tgz. Deprecated in favour of [kobo-fwinfo](../kobo-fwinfo/).

````
Usage: kobo-versionextract [OPTIONS] PATH_TO_FW

Options:
  -h, --help   show this help text

PATH_TO_FW is either the path to a KoboRoot.tgz, libnickel.so, or a kobo update zip.
Note that kobo-versionextract only works with firmware 4.7.10413 and later.
````