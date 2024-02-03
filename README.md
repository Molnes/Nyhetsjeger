# Nyhetsjeger  
Bachelor Project for SMP  

# Tools  
In order to run and develop this project, you will need a set of tools.  

- [Golang](https://go.dev/) - the project is primarily written in GO  
- [Templ](https://templ.guide/quick-start/installation) - HTML templating language for Go  
    you need it installed in order to generate go files from templates  
    - You will also need [templ-vscode](https://marketplace.visualstudio.com/items?itemName=a-h.templ) extension, for improved developer experience.  
    - For better IDE support, Tailwind completions etc, check out [Templ's guide](https://templ.guide/commands-and-tools/ide-support/)  
    - For HTMX suggestions in Templ files, you may want to use [HTMX Attributes](https://marketplace.visualstudio.com/items?itemName=CraigRBroughton.htmx-attributes) extension.  
- [Air](https://github.com/cosmtrek/air?tab=readme-ov-file#via-go-install-recommended) - Live reload for Go apps  
    improved developer experience  
- [Node](https://nodejs.org/) to install and run [Tailwind CSS](https://tailwindcss.com/).  
    - You will need [Tailwind CSS IntelliSense](https://marketplace.visualstudio.com/items?itemName=bradlc.vscode-tailwindcss) vscode extension for Tailwind suggestions.  

# Building and running locally  
Use **Make** to run targets (build, run, etc.) specified in the [Makefile](./Makefile). Make is included in most Linux distributions(as well as WSL).  
Example:  
```bash
make run
```

Alternatively, you can check out the commands in the [Makefile](./Makefile) and run them manually.
