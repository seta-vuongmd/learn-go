package graphql

import (
	"errors"
	"user-team-asset-management/internal/auth"
	"user-team-asset-management/internal/models"
	"user-team-asset-management/internal/utils"

	"github.com/graphql-go/graphql"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Resolver struct {
	DB        *gorm.DB
	JWTSecret string
}

func (r *Resolver) CreateSchema() (graphql.Schema, error) {
	userType := graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"userId":   &graphql.Field{Type: graphql.String},
			"username": &graphql.Field{Type: graphql.String},
			"email":    &graphql.Field{Type: graphql.String},
			"role":     &graphql.Field{Type: graphql.String},
		},
	})

	loginResponseType := graphql.NewObject(graphql.ObjectConfig{
		Name: "LoginResponse",
		Fields: graphql.Fields{
			"token": &graphql.Field{Type: graphql.String},
			"user":  &graphql.Field{Type: userType},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"fetchUsers": &graphql.Field{
				Type:    graphql.NewList(userType),
				Resolve: r.fetchUsers,
			},
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createUser": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"username": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"email":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"password": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"role":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: r.createUser,
			},
			"login": &graphql.Field{
				Type: loginResponseType,
				Args: graphql.FieldConfigArgument{
					"email":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"password": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: r.login,
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
}

func (r *Resolver) createUser(p graphql.ResolveParams) (interface{}, error) {
	username := p.Args["username"].(string)
	email := p.Args["email"].(string)
	password := p.Args["password"].(string)
	role := p.Args["role"].(string)

	if role != "manager" && role != "member" {
		return nil, errors.New("role must be 'manager' or 'member'")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID := utils.GenerateID()
	user := models.User{
		ID:           userID,
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         role,
	}

	if err := r.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Resolver) login(p graphql.ResolveParams) (interface{}, error) {
	email := p.Args["email"].(string)
	password := p.Args["password"].(string)

	var user models.User
	if err := r.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := auth.GenerateToken(user.ID, user.Role, r.JWTSecret)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"token": token,
		"user":  user,
	}, nil
}

func (r *Resolver) fetchUsers(p graphql.ResolveParams) (interface{}, error) {
	var users []models.User
	if err := r.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
