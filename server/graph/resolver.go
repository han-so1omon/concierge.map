package graph

//go:generate rm -rf generated
//go:generate go run github.com/99designs/gqlgen --verbose

import (
	"context"
	"errors"
	"github.com/han-so1omon/concierge.map/data/model"
	"github.com/han-so1omon/concierge.map/server/graph/generated"
	"github.com/han-so1omon/concierge.map/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (q *queryResolver) Projects(ctx context.Context) ([]*model.Project, error) {
	cursor, err := q.ProjectCollection.Find(ctx, bson.D{})

	if err == mongo.ErrNoDocuments {
		return nil, errors.New("no existing projects")
	}
	if err != nil {
		return nil, errors.New("internal error")
	}

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, errors.New("internal error")
	}

	var projects []*model.Project
	for _, result := range results {
		project := model.Project{
			ID:          result["_id"].(primitive.ObjectID).String(),
			Name:        result["name"].(string),
			Description: result["description"].(string),
			OwnerID:     result["ownerId"].(string),
		}
		if err != nil {
			return nil, errors.New("internal error")
		}
		projects = append(projects, &project)
	}

	return projects, nil
}

func (q *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	cursor, err := q.UserCollection.Find(ctx, bson.D{})

	if err == mongo.ErrNoDocuments {
		return nil, errors.New("no existing users")
	}
	if err != nil {
		return nil, errors.New("internal error")
	}

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, errors.New("internal error")
	}

	var users []*model.User
	for _, result := range results {
		//TODO figure out how to properly unmarshal primitive.A
		var projectIds []string
		for _, projId := range result["projectIds"].(primitive.A) {
			projectIds = append(projectIds, projId.(primitive.M)["_id"].(string))
		}
		user := model.User{
			ID:            result["_id"].(primitive.ObjectID).String(),
			Username:      result["username"].(string),
			Email:         result["email"].(string),
			Password:      result["password"].(string),
			Authenticated: result["authenticated"].(bool),
			ProjectIDs:    projectIds,
		}
		users = append(users, &user)
	}

	return users, nil

}

func (m *mutationResolver) CreateProject(ctx context.Context, input model.NewProject) (*model.Project, error) {
	if input.Name == "" {
		return nil, nil
	}

	var existingProject model.Project
	// Check if project already exists
	mongoErr := m.ProjectCollection.FindOne(ctx, bson.M{
		"name":        input.Name,
		"description": input.Description,
	}).Decode(&existingProject)

	if mongoErr != nil && mongoErr != mongo.ErrNoDocuments {
		return nil, errors.New("internal error")
	}

	if existingProject.Name == input.Name {
		return nil, errors.New("project with this name and description already exists")
	}

	var user, updatedUser model.User
	err := m.UserCollection.FindOne(ctx, bson.M{
		"email": input.OwnerEmail,
	}).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return nil, errors.New("project owner does not exist")
	}
	if err != nil {
		return nil, errors.New("internal error")
	}

	_, err = m.ProjectCollection.InsertOne(ctx, bson.M{
		"name":        input.Name,
		"description": input.Description,
		"ownerId":     user.ID,
	})

	if err != nil {
		return nil, errors.New("internal error")
	}

	var project model.Project
	err = m.ProjectCollection.FindOne(ctx, bson.M{
		"name":        input.Name,
		"description": input.Description,
		"ownerId":     user.ID,
	}).Decode(&project)
	if err != nil {
		return nil, errors.New("internal error")
	}

	opts := options.FindOneAndUpdate().SetUpsert(false)
	err = m.UserCollection.FindOneAndUpdate(
		ctx,
		bson.D{{Key: "email", Value: user.Email}},
		bson.D{{Key: "$push", Value: bson.D{{Key: "projectIds", Value: bson.D{{Key: "_id", Value: project.ID}}}}}},
		opts,
	).Decode(&updatedUser)
	if err != nil {
		return nil, errors.New("internal error")
	}

	return &project, nil
}

func (m *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (*model.User, error) {
	if input.Email == "" {
		return nil, nil
	}
	var existingUser bson.M
	// Check if email already exists
	mongoErr := m.UserCollection.FindOne(ctx, bson.M{
		"email": input.Email,
	}).Decode(&existingUser)

	if mongoErr != nil && mongoErr != mongo.ErrNoDocuments {
		util.Logger.Info(mongoErr)
		return nil, errors.New("internal error")
	}
	if existingUser["email"] == input.Email {
		return nil, errors.New("user with email already exists")
	}

	res, err := m.UserCollection.InsertOne(ctx, bson.M{
		"username":      input.Username,
		"email":         input.Email,
		"password":      input.Password,
		"authenticated": false,
		"projectIds":    []string{},
	})

	if err != nil {
		return nil, errors.New("internal error")
	}

	var newUser model.User
	err = m.UserCollection.FindOne(ctx, bson.M{
		"_id": res.InsertedID,
	}).Decode(&newUser)

	if err != nil {
		return nil, errors.New("internal error")
	}

	return &newUser, nil
}

func (p *projectResolver) Owner(ctx context.Context, obj *model.Project) (*model.User, error) {
	//var user model.User
	objId, err := primitive.ObjectIDFromHex(obj.OwnerID)
	var userM bson.M
	err = p.UserCollection.FindOne(ctx, bson.M{
		"_id": objId,
	}).Decode(&userM)

	var projectIds []string
	for _, projId := range userM["projectIds"].(primitive.A) {
		projectIds = append(projectIds, projId.(primitive.M)["_id"].(string))
	}
	user := model.User{
		ID:            userM["_id"].(primitive.ObjectID).String(),
		Username:      userM["username"].(string),
		Email:         userM["email"].(string),
		Authenticated: userM["authenticated"].(bool),
		Password:      userM["password"].(string),
		ProjectIDs:    projectIds,
	}

	if err == mongo.ErrNoDocuments {
		return nil, errors.New("user with id does not exist")
	}
	if err != nil {
		return nil, errors.New("internal error")
	}

	return &user, nil
}

func (u *userResolver) Projects(ctx context.Context, obj *model.User) ([]*model.Project, error) {
	if len(obj.ProjectIDs) == 0 {
		return []*model.Project{}, nil
	}

	projectIds := make([]primitive.ObjectID, 0, len(obj.ProjectIDs))

	for _, projIdStr := range obj.ProjectIDs {
		pid, pidErr := primitive.ObjectIDFromHex(projIdStr)
		if pidErr != nil {
			return nil, errors.New("internal error")
		}

		projectIds = append(projectIds, pid)
	}

	cursor, err := u.ProjectCollection.Find(ctx, bson.D{{
		Key: "_id", Value: bson.D{{Key: "$in", Value: projectIds}},
	}})
	if err != nil {
		return nil, errors.New("internal error")
	}

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, errors.New("internal error")
	}

	var projects []*model.Project
	for _, result := range results {
		project := model.Project{
			ID:          result["_id"].(primitive.ObjectID).String(),
			Name:        result["name"].(string),
			Description: result["description"].(string),
			OwnerID:     result["ownerId"].(string),
		}
		if err != nil {
			return nil, errors.New("internal error")
		}
		projects = append(projects, &project)
	}

	return projects, nil
}

func (r *Resolver) Query() generated.QueryResolver {
	return &queryResolver{r}
}
func (r *Resolver) Mutation() generated.MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Project() generated.ProjectResolver {
	return &projectResolver{r}
}
func (r *Resolver) User() generated.UserResolver {
	return &userResolver{r}
}

type Resolver struct {
	UserCollection    *mongo.Collection
	ProjectCollection *mongo.Collection
}
type queryResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type projectResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
