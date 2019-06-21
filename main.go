package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/rs/cors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func main() {
	// Create a handler for /graphql which passes cors for remote requests
	http.Handle("/graphql", cors.Default().Handler(&relay.Handler{Schema: graphqlSchema}))

	// Write a GraphiQL page to /
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(page)
	}))

	// ListenAndServe starts an HTTP server with a given address and handler.
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func init() {

	// MustParseSchema parses a GraphQL schema and attaches the given root resolver.
	// It returns an error if the Go type signature of the resolvers does not match the schema.
	graphqlSchema = graphql.MustParseSchema(Schema, &Resolver{})
}

func GetMongo(col string) (*mgo.Session, *mgo.Collection) {
	maxWait := time.Duration(5 * time.Second)
	session, err := mgo.DialWithTimeout("localhost", maxWait)

	if err != nil {
		log.Fatal(err)
	}

	collection := session.DB("graphql1").C(col)

	return session, collection
}

func Cleanup(col string) {
	log.Println("Cleaning up MongoDB...")
	session, _ := GetMongo(col)
	defer session.Close()
}

var graphqlSchema *graphql.Schema

var Schema = `
    schema {
        query: Query
    }
    # The Query type represents all of the entry points.
    type Query {
			user(city: String!): User
			post(slug: String!): Post
    }
    type User {
			  name: String!
        age: Int!
        city: String!
		}
		type Post {
			id: ID!
			slug: String!
			title: String!
		}
    `

type Resolver struct{}

type post struct {
	ID    graphql.ID
	Slug  string
	Title string
}

type postResolver struct {
	s *post
}

type user struct {
	Name string
	Age  int32
	City string
}

type userResolver struct {
	s *user
}

type searchResultResolver struct {
	result interface{}
}

var userData = make(map[string]*user)

// User resolves the User queries.
func (r *Resolver) User(args struct{ City string }) *userResolver {

	// One result is a pointer to type user.
	// oneResult := &user{}
	oneResult := &user{}

	session, collection := GetMongo("user")
	// Close the session so its resources may be put back in the pool or collected, depending on the case.
	defer session.Close()

	// Inside the collection, find by city and return all fields.
	// err := collection.Find(bson.M{"slug": args.Slug}).Select(bson.M{}).One(&oneResult)
	err := collection.Find(bson.M{"city": args.City}).Select(bson.M{}).One(&oneResult)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(oneResult)

	//Make a type postResolver out of oneResult.
	if s := oneResult; s != nil {
		return &userResolver{oneResult}
	}
	return nil
}

func (r *Resolver) Post(args struct{ Slug string }) *postResolver {

	// One result is a pointer to type user.
	oneResult := &post{}

	session, collection := GetMongo("post")
	print(collection.Find(bson.M{}))
	// Close the session so its resources may be put back in the pool or collected, depending on the case.
	defer session.Close()

	// Inside the collection, find by city and return all fields.
	err := collection.Find(bson.M{"slug": args.Slug}).Select(bson.M{}).One(&oneResult)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(oneResult)

	// //Make a type postResolver out of oneResult.
	if s := oneResult; s != nil {
		return &postResolver{oneResult}
	}
	return nil
}

// Resolve each field to respond to queries.
func (r *userResolver) Name() string {
	return r.s.Name
}

func (r *userResolver) Age() int32 {
	return r.s.Age
}

func (r *userResolver) City() string {
	return r.s.City
}

func (r *postResolver) ID() graphql.ID {
	return r.s.ID
}

func (r *postResolver) Slug() string {
	return r.s.Slug
}

func (r *postResolver) Title() string {
	return r.s.Title
}

var page = []byte(`
    <!DOCTYPE html>
    <html>
        <head>
            <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.10.2/graphiql.css" />
            <script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/1.1.0/fetch.min.js"></script>
            <script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react.min.js"></script>
            <script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react-dom.min.js"></script>
            <script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.10.2/graphiql.js"></script>
        </head>
        <body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
            <div id="graphiql" style="height: 100vh;">Loading...</div>
            <script>
                function graphQLFetcher(graphQLParams) {
                    return fetch("/graphql", {
                        method: "post",
                        body: JSON.stringify(graphQLParams),
                        credentials: "include",
                    }).then(function (response) {
                        return response.text();
                    }).then(function (responseBody) {
                        try {
                            return JSON.parse(responseBody);
                        } catch (error) {
                            return responseBody;
                        }
                    });
                }
                ReactDOM.render(
                    React.createElement(GraphiQL, {fetcher: graphQLFetcher}),
                    document.getElementById("graphiql")
                );
            </script>
        </body>
    </html>
    `)
