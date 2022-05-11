# Testing tool for Arduino libraries 🗜

This (unofficial) tool performs batch tests on the libraries indexed in the [Arduino Library Registry](https://github.com/arduino/library-registry). See the [published HTML report](https://alranel.github.io/arduino-testlib/) for an example of the output you can generate.

For each library_version/core_version pair, the following tests will be done:

* Compilation of an empty sketch that only includes the main header file (eg. `#include <Servo.h>` for the Servo library). This will check that the header file itself compiles, as well as any other .cpp file included in the library. If no main header file exists, an empty one is generated to allow at least the compilation of .cpp files.
* Compilation of all the examples included in the library. Since examples often require libraries that are not specified as dependencies in library.properties, it is recommended to have all the existing libraries installed locally before starting the test.
* Compliance check between the test results and the supported architectures declared in the [library.properties](https://arduino.github.io/arduino-cli/0.20/library-specification/) metadata file.

Notes and limitations:

* other header files that are distributed with the library but are not included by the main header file or a .cpp file or by the example sketches will not be tested for compilation;
* successful compilation for a given board does not guarantee full compatibility because there could be runtime issues or specific hardware may be required.

## Getting started

To get started, compile the tool with a simple `go build`.

### Testing all the Arduino libraries

1. Install/upgrade all the libraries indexed in the Arduino Library Registry:
    * `./arduino-testlib installall --cli-datadir path/to/dir`
2. Test all the libraries:
    * `./arduino-testlib testall --cli-datadir path/to/dir --datadir path/to/dir --fqbn arduino:avr:uno`
3. Generate an HTML report:
    * `./arduino-testlib report --datadir path/to/dir`

Available options:

* `--cli-datadir`: a local directory that will be used to store your libraries and platforms without polluting your default arduino-cli setup. May be omitted but it's highly recommended. Just create an empty directory and point to it.
* `--datadir`: a local directory that will be used to store the JSON files with the test results of each library
* `--threads`: this can be used in combination with the `testall` command to parallelize tests
* `--fqbn`: use this option to specify the boards to test with; can be used multiple times
* `--force`: use this with `testall` to force testing of library_version/core_version that were already seen; if not specified, they will be skipped to allow incremental runs

### Testing individual libraries

This tool can be also used to test a specific library. You can think about it as a wrapper around `arduino-cli compile` that will try to run all the possible compilation tests for a given library and print the result.

```
./arduino-testlib test --fqbn arduino:avr:uno path/to/lib
```

## Credits and license

This tool was written by [Alessandro Ranellucci](https://github.com/alranel) and is licensed under the terms of the Affero GNU General Public License v3.