package main

// i try to put comment on every line to code so fellow programmer can understand this code easily ! ;-)
import (
	"context"
	"fmt"     //format
	"log"     // its simple showw logs like clg inm node js
	"os"      // load env in our case
	"strconv" // convert value
	"time"    // current time

	"github.com/gofiber/fiber/v2" // fiber
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv" //insall the env packages
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive" // for define thee primitive object id
	"go.mongodb.org/mongo-driver/mongo"          // install for mongodb access
	"go.mongodb.org/mongo-driver/mongo/options"
)

// kind of type Script thing
type Movie struct{
	Id primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title string `json:"title"`
	Genre string `json:"genre"`
	Year int `json:"year"`
	Rating int `json:"rating"`
	
}

// collection Schema access
var collection *mongo.Collection  

func main() {
	// just check working or not
	fmt.Println("Start")
	
	//Task1: Load .env file start here
	err := godotenv.Load(".env") // path of env file
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err.Error())
	}
	fmt.Println("Environment variables loaded successfully")
	// Load .env file end here now we can use enviromentt varibales  in our app


	//Task2: make a dataBase connection start here 
	    //  - install mongo driver by running this command: go get go.mongodb.org/mongo-driver/mongo
		//  - Establish the conncetions

		//1. import the mongoDB uri from .env
		MongoDB_uri := os.Getenv("MONGO_URI");
		fmt.Println(MongoDB_uri)
	
		//2. Make a connection using Client Options
		clientOptions := options.Client().ApplyURI(MongoDB_uri);


		//3. Establish Connection
		client,err := mongo.Connect(context.Background(),clientOptions);
		if err != nil{
			log.Fatalf("Failed to connect with MongoDb with error: %v",err.Error());
		}
		
		// end the connection when the some error occur using defer 
		defer client.Disconnect(context.Background());

		//4. ping to verify the connection is establish successfully or not 
		err = client.Ping(context.Background(),nil)
		if err != nil {
			log.Fatalf("Failed to ping MongoDB: %v", err.Error())
		}

		fmt.Println("Successfully Connect to MongoDB...");

		//5. Initalise the collection from database 

		collection = client.Database("go_langDatabase").Collection("movie");
		fmt.Println("Collection instance created:", collection.Name());

	// Task 2: Use of fiber , it's like Express framework that use to make the end point in go
		
		//1. initilization of fiber in our app by run this command :go get github.com/gofiber/fiber/v2

		app:= fiber.New();


		//DONE - cors
		app.Use(cors.New(cors.Config{
			AllowOrigins: "*", // Allow all origins for development
			AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
			AllowHeaders: "Origin,Authorization,Content-Type,Accept",
		}))


		frontendURL := os.Getenv("FRONTEND_URL");
		if frontendURL == "" {
			frontendURL = "http://localhost:3000" 
		}
		// please  ingnore this its my practice where i check the port and all this is up or not 
	
		

// 		1.) Create/Update Movie Detail , created - done cheked , updated - done
// - Title - done 
// - Genre - done 
// - Year - done
// - Rating (max 5) - done checked 

			// now define our all Routes here  check controllers at bottom
			// 1 get all moviesss from the database  with pagination  and search according to genre and year 

			app.Get("/getmovies",getAllMovies);

			//2. for create a movie usiing the Post req { title,Genre,year,rating}
			app.Post("/createmovie",createMovie);

			//3. for update movie using the put req {all optional}
			app.Put("/updatemovie/:id",updateMovie);

			//4. for delete existing movie use DELETE req 
			app.Delete("/deletemovie/:id",deleteMovie);

			//4. for find single existing movie use get req 
			app.Get("/singlemovie/:id",getMovieByID);

// ## Validation
// -- should not have duplicate name - done
// --  year should be between 1900 and current year only. - done

// 2.) List
// - should be listed as tile view - Title, Genre, Year - done
// - should show rating with star. -done
// - should have delete icon to delete movie - done
// - should be able to search movie using Title (implement in API) - done
// - should be able to filter movie using Genre and Year (implement in API)  - done

// Extra Addon : Add pagination to list view - done



		
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" 
	}
	fmt.Printf("Listening on http://localhost:%v\n", port)


	// Start the server
	log.Fatal(app.Listen("0.0.0.0:" + port))


}



// function to get all movies with pagination 
func getAllMovies(c *fiber.Ctx) error{
	 var movies  []Movie; //this is Slice that hold multiple movies [{},{}]
	 title := c.Query("title")
	 pageStr := c.Query("page","1") // page default value as 1
	 limitStr := c.Query("limit","10");
	 genre := c.Query("genre"); //#regex 
	 yearStr := c.Query("year");

	 fmt.Println(pageStr ,limitStr,genre,yearStr); //reciving value success "http://localhost:5000/getmovies?page=1&limit=10&genre=comedy&year=2000"
	

	//  Convert the string value to number 
	// check if the number of page < 0 ,error found in case -> then assgin the value by default zero
	 page,err:= strconv.Atoi(pageStr);
	 if err != nil || page < 0{
		page = 1;
	 }

	// above thing same done here
	// we can use use uint instead of int datatype that only recives none negative integers here 
	limit,err := strconv.Atoi(limitStr);
	if err != nil || limit < 0 {
		limit =10
	}

	skip := (page-1) * limit ;
	fmt.Println(skip);  // 1-1 = 0 *10 =0  // 2-1 =1*10 at page second it skip the 10 values from starting 


	filter := bson.M{} //create a empty bson Map for filter movies according to genre 
	if genre != ""{ //if genre is available then it pput the value in the map 
		filter["genre"] = bson.M{
			"$regex":   genre,
			"$options": "i", 
		}
	}
	if title != "" {
		filter["title"] = bson.M{
			"$regex":   title,
			"$options": "i", 
		}
	}
	if yearStr != ""{
		year,err := strconv.Atoi(yearStr);
		if err == nil {
			filter["year"] = year
		}
	}

	findOptions := options.Find();
	findOptions.SetLimit(int64(limit));
	findOptions.SetSkip(int64(skip));

	// Count the total number of documents that match the filter
	totalCount, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count movies"})
	}

	 cursor,err := collection.Find(context.Background(),filter,findOptions);
	 if err != nil {
		return c.Status(500).JSON(fiber.Map{"error":"Failed to fetch movies"})
	 }
	 defer cursor.Close(context.Background()); // when function close , connection close 
	 

	 for cursor.Next(context.Background()){
		var movie Movie;
		if err := cursor.Decode(&movie); err!=nil{
			return c.Status(500).JSON(fiber.Map{"error":"Error Occur in Decode Movie","message":err.Error()})
		}
		movies = append(movies, movie);
	 }

	 totalPages := (totalCount + int64(limit) - 1) / int64(limit)
	 

	 
	
	return c.Status(200).JSON(fiber.Map{"success":"true","data":movies,"totalPages":totalPages,"totalCount":totalCount})
}


//function for creating a movie

func createMovie(c *fiber.Ctx) error{
	movie := new(Movie) // Xcreate a new Struct 
	if err := c.BodyParser(movie); err !=nil{
		fmt.Println("Body Pareser with err: ",err);
		return c.Status(400).JSON(fiber.Map{"success":false,"message":err.Error()});
	}
	fmt.Println(movie.Title , movie.Genre ,movie.Year)
	
	// Validate the fields
	if movie.Title == "" {
        return c.Status(400).JSON(fiber.Map{"success": false, "message": "Title is required"})
    }
    if movie.Genre == "" {
        return c.Status(400).JSON(fiber.Map{"success": false, "message": "Genre is empty","error":"genre cannot be empty"})
    }
	if movie.Rating <0 || movie.Rating > 5 {
        return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid rating","error": "rating should be between 0 and 5"})
    }
    if movie.Year <= 0 || movie.Year < 1900 || movie.Year > time.Now().Year() {
        return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid year", "error": "Year should be between 1900 and the current year"})
    }

	// check duplicate movie exist or not before inserting
	var ExistMovie = new(Movie);
	filter := bson.M{"title":movie.Title};
	err := collection.FindOne(context.Background(),filter).Decode(&ExistMovie);
	if err ==nil{
		return c.Status(500).JSON(bson.M{"success":false,"message":"duplicate entry","error":"Movie Already exists!"});
	}



	// insert into Our movie collection
	
	insertResult , err := collection.InsertOne(context.Background(),movie)
	if err != nil{
		return c.Status(500).JSON(bson.M{"success":false,"message":"faliled to create movie","error":err.Error()});
	}
	movie.Id = insertResult.InsertedID.(primitive.ObjectID)
	return c.Status(200).JSON(fiber.Map{"success":true,"message":"update Successfully","data":movie});
}

//funxtion for updating movie
func updateMovie(c *fiber.Ctx) error{
	movieid := c.Params("id");
	fmt.Println(movieid); //recived succeess check 
	
	// convert id string to objectid
	objectId ,err := primitive.ObjectIDFromHex(movieid);
	if err !=nil {
		return c.Status(500).JSON(fiber.Map{"success":false ,"message":"invalid id","error":err.Error()})
	}


	updated := bson.M{} //slice for handle updated fields


	// check moviw exists or not if not exist then simpley return in early process

	filter:= bson.M{"_id":bson.M{"$eq":objectId}}
	count,err := collection.CountDocuments(context.Background(),filter);

	if err !=nil{
		return c.Status(500).JSON(fiber.Map{"success":false,"message":"error occur in check movie exist!","error":err.Error()});
	}
	
	if count ==0{
		return c.Status(404).JSON(fiber.Map{"success":false,"message":"missing item","error":"Movie not found!"});
	}

	// get the Request body
	movie := new(Movie);
	if err := c.BodyParser(movie) ;err != nil{
		return c.Status(500).JSON(fiber.Map{"success":false})
	}

	if movie.Title !="" {
	//1. check movie titile must be unique
	filter := bson.M{"_id":bson.M{"$ne":objectId},"title":movie.Title};
	count,err:= collection.CountDocuments(context.Background(),filter);
	if err != nil{
		return c.Status(500).JSON(fiber.Map{"success":false,"message":err.Error(),"error":"Error Occur in cheking for duplicate Id"});
	}

	//if already exist then throws error ...
	if count > 0 {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Movie with this title already exists!!!",})
	}
	
	//if all passes the validations
	updated["title"] = movie.Title // update the Slice
	}

	if movie.Genre != "" {
		updated["genre"] = movie.Genre
	}

	if movie.Year > 0 {
		if movie.Year < 1900 || movie.Year > time.Now().Year() {
			return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid year", "error": "Year should be between 1900 and the current year"})
		}
		updated["year"] = movie.Year
	}

	if movie.Rating >= 0 && movie.Rating <= 5 {
		updated["rating"] = movie.Rating
	} else if movie.Rating < 0 || movie.Rating > 5 {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid rating", "error": "Rating should be between 0 and 5"})
	}

	if len(updated) == 0 {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "No fields to update"})
	}


	// Add updated_at field
	updated["updated_at"] = time.Now()


	// finally update the movie
	update_filter := bson.M{"$set":updated};
	_,err = collection.UpdateByID(context.Background(),objectId,update_filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to update movie", "message": err.Error()})
	}
	return c.Status(200).JSON(fiber.Map{"success":true,"message":"updated successfully!!"});
}


// function for Deleting movie
func deleteMovie(c *fiber.Ctx) error{
	id := c.Params("id");
	objectId ,err := primitive.ObjectIDFromHex(id);
	
	filter := bson.M{"_id":objectId};
	if err != nil{
		return c.Status(400).JSON(fiber.Map{"success":false,"error": "Invalid ID","message":err.Error()})
	}
	
	// delete document
	_,err = collection.DeleteOne(context.Background(),filter);
	if err != nil {
        return c.Status(500).JSON(fiber.Map{"success":false,"error": "Error deleting data","message":err.Error()})
    }
	return c.Status(200).JSON(fiber.Map{"success": true, "message":"Deleted Successfully"})
}


func getMovieByID(c *fiber.Ctx) error {
	id := c.Params("id") // Get the ID from the route parameters

	// Convert the id string to an ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid movie ID"})
	}

	var movie Movie
	filter := bson.M{"_id": objID}

	err = collection.FindOne(context.Background(), filter).Decode(&movie)
	if err == mongo.ErrNoDocuments {
		return c.Status(404).JSON(fiber.Map{"error": "Movie not found"})
	} else if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch movie", "message": err.Error()})
	}

	return c.Status(200).JSON(fiber.Map{"success": "true", "data": movie})
}