## Booking and Reservation Project

This is my first fullstack Go project. Built with Go version 1.19

This project is a simple hotel booking and reservation project with key features like;

- Authentication
- Showcase the properties
- Check for rooms availability
- Allow booking of one room
- Notify guest and owner
- Have a backend for management
- Review existing bookings
- Show a calendar of bookings
- Change or cancel a booking

### Packages used

- [Alex Edwards SCS](https://github.com/alexedwards/scs/v2) package for managing sessions
- [Chi router](https://github.com/go-chi/chi/v5)
- [Justinas nosurf](https://github.com/justinas/nosurf)

### Note: 
- Create your own go mod file and delete the one used here, run the following command `go mod init your-project-name`
- your-project-name is usually your github link and the name of your project, example "github.com/prosper74/go-project". This is not a must, but a recommendation.
- Change the name of every import to your current go mod name. Example, open the main.go file, in the `required imports` section, replace these "github.com/atuprosper/go-project/pkg/config" to "github.com/atuprosper/your-project-name/pkg/config". Go through all files and make this replacement
- After all the necessary changes, run the app `go run cmd/web/*.go` this will install all the third party packages and run the server on the selected port.

### Run the server
- Manual: `go run cmd/web/main.go cmd/web/middleware.go cmd/web/routes.go`

### The test file 
- To output test in hmtl format run `go test -coverprofile=coverage.out && go tool cover -html=coverage.out`
- To know the percentage coverage run `go test -cover`

### The main.go file
This is where we create and configure our session

- The seesion variable `var session *scs.SessionManager` here and point (*) it to the scs package. This will make the session variable available to other files in this project.

- Here `session = scs.New()` we assign a new scs session to the sessions variable.
	
- `session.Lifetime = 24 * time.Hour` states how long will the sessions last. `time.Hour` is a Go built in package 
	
- `session.Cookie.Persist = true` Here we store the session in a cookie. We can also use other databases like redis to store our session. This will make the session persist for this time duration even if the browser window is closed and openned again. If its set to false, the session ends when the brower is closed
	
- To set the strictness of the session, we use `session.Cookie.SameSite = http.SameSiteLaxMode`
	
- To make sure the connection is encrypted (i.e https), we use `session.Cookie.Secure = app.InProduction`. In production, this should be true. In development, set it to false. Note that `app.InProduction` is set in the site wide `config.go` file

- Finally, we make our session available site wide by passing it to `app.Session = session`

### The handlers.go file
This where we handle all our pages and templates.
This is where we use our scs package and all it's features. We want to collect the IP address of our users and store them in the session, then send them to be used in our HTML templates

- `func (m *Repository) Home(w http.ResponseWriter, r *http.Request)` A function with a 'reciever' m, of type 'Repository'. This will give our handler function access to everything in the config file
- `var Repo *Repository` Creating a Repository pattern. This variable is the repository used by the handlers
- Get the IP address of user using the r parameter, which holds the `*http.Request` Go built in package - `remoteIP := r.RemoteAddr`
- Store the IP in the site wide config which is available via the paramter 'm'. It takes in three parameters. The context, the name (which can be anything), and the value - `m.App.Session.Put(r.Context(), "remote_ip", remoteIP)`
- Get the remote IP from the session - `getRemoteIP := m.App.Session.GetString(r.Context(), "remote_ip")`
- Store the retrieved IP in the `stringMap` which is available in our HTML template - `stringMap["remote_ip"] = getRemoteIP`

### The routes.go file
This is where we handle routes, multiplexers and middlewares 

- `mux := chi.NewRouter()` Creates a multiplexer
Make the templates recognize the static folder to use its contents (eg images)
- Create a file server for the static directory `fileServer := http.FileServer(http.Dir("./static/"))`
- Create a multiplexer to handle the fileServer. "http.StripePrefix" takes the url and modifies it to something it can handle. `mux.Handle("/static/*", http.StripPrefix("/static", fileServer))`
- 'Recoverer' middleware from chi package. It helps for panic control `mux.Use(middleware.Recoverer)`

### The render.go file 
- `template.FuncMap{}` A FuncMap is a map of functions that we can use in our template
- `NewTemplates` sets the config for the template package
- `AddDefaultData` This function adds default data to every templates by taking the specific data sent from the render function and attach this function to it

### The middleware.go file 
We can write our own middleware for `chi` package in this file.

- `next` is commonly used as the parameter for custome middleware
- `NoSurf` middleware adds CSRF protection to all POST request - `func NoSurf(next http.Handler) http.Handler`
- `SessionLoad` middleware makes our server to be 'state' aware, in order to keep our session in state `func (next http.Handler) http.Handler`
- `LoadAndSave` provides middleware which automatically loads and saves session data for the current request, and communicates the session token to and from the client in a cookie.

### The config.go file 
This file should be made accessible by other packages, every package in this project can import it and use it. But this file should not import any package, it will only use the standard built-in Go packages. This will help us avoid an error called "import cycle"

- `AppConfig` struct holds the application configurations
- `InfoLog` allows you to create a log file and store informations in it

### The templateData.go file 
We create this seperate file to avoid Go `import cycle not allowed error`

- `TemplateData struct` is a struct that holds data sent from handlers to templates
- `interface{}` type is used when the type is not known
- Cross Site Request Forgery Token (CSRFToken) - is a security token to handle forms

### The test file 
- To output test in hmtl format run `go test -coverprofile=coverage.out && go tool cover -html=coverage.out`
- To know the percentage coverage run `go test -cover`
