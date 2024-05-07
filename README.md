# Nyhetsjeger  
Bachelor Project for SMP  

# Tools  
In order to run and develop this project, you will need a set of tools.  

- [Visual Studio Code](https://code.visualstudio.com/) - editor recommanded for this project, tools and plugins listed below transform it into an IDE.

- [WSL](https://learn.microsoft.com/en-us/windows/wsl/install) - We use it for consistant developer experience, as well as development in similar environment to production. Allows us to use Linux-specific toolchains. This simplifies development of scripts.  

**All tooling below must be accessible in WSL**

- [Golang](https://go.dev/doc/install) - the project is primarily written in GO  
    Remember to add go and go's binaries to your **path**!  
    file: `~/.bashrc` (or other shell's configuration file), add the following lines at the end of it
    ```bash
    export PATH=$PATH:/usr/local/go/bin
    export PATH=$PATH:$HOME/go/bin
    ```
- [Templ](https://templ.guide/quick-start/installation) - HTML templating language for Go  
    you need it installed in order to generate go files from templates  
    - You will also need [templ-vscode](https://marketplace.visualstudio.com/items?itemName=a-h.templ) extension, for improved developer experience.  
    - For better IDE support, Tailwind completions etc, check out [Templ's guide](https://templ.guide/commands-and-tools/ide-support/).  
    - For HTMX suggestions in Templ files, you may want to use [HTMX Attributes](https://marketplace.visualstudio.com/items?itemName=CraigRBroughton.htmx-attributes) extension.  
- [Air](https://github.com/cosmtrek/air?tab=readme-ov-file#via-go-install-recommended) - Live reload for Go apps  
    improved developer experience  
- [Node](https://nodejs.org/) to run [Tailwind CSS](https://tailwindcss.com/).
    -  **Recommended** installation via [Node Version Manager](https://github.com/nvm-sh/nvm) - otherwise WSL might try to run Windows' Node installation, and this **will** break. 
    - You will need [Tailwind CSS IntelliSense](https://marketplace.visualstudio.com/items?itemName=bradlc.vscode-tailwindcss) vscode extension for Tailwind suggestions.  
- [Docker](https://docs.docker.com/get-docker/) - for running the database locally and deploying the app
    - WSL will need **Docker Desktop for Windows** anyway, usually works out the box. For troubleshooting see Docker's and Microsoft's documentation
- [migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) - SQL migration tool
    simplest way to install
    ```bash
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    ```
# Building and running locally  
Use **Make** to run targets (`build`, `run`, etc.) specified in the [Makefile](./Makefile). Make is included in most Linux distributions(as well as WSL). If it isnt included by default, use your distros package manager to install it.  
Example:  
```bash
make run
```
Alternatively, you can check out the commands in the [Makefile](./Makefile) and run them manually.



# Testing
The project has unit and integration tests, they live in separate files and are tagged with go build tags.  


**unit tests**: `filename_test.go`, build tag `unit`  
**integration tests**: `filename_integration_test.go`, build tag `integration`  

Add the build tags to your LSP configuration so you get syntax highlighting and autocompletion in the test files.  
VScode `settings.json` example:  
```json
{
    "go.buildFlags": [
        "-tags=integration,unit"
    ],
    "go.testTags": "integration,unit",
}
```

To run the tests, use specific Make targets
```bash
make test-unit
make test-integration
```

Note: Integration tests use [Testcontainers](https://golang.testcontainers.org/), which requires Docker to be running.