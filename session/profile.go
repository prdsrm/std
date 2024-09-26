package session

import (
	"context"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
)

// DeleteProfilePictures delete all current user profile picture.
func DeleteProfilePictures(ctx context.Context, client *telegram.Client) error {
	self, err := client.Self(ctx)
	if err != nil {
		return err
	}
	elems, err := query.NewQuery(client.API()).GetUserPhotos(self.AsInput()).Collect(ctx)
	if err != nil {
		return err
	}
	var inputPhotos []tg.InputPhotoClass
	for _, elem := range elems {
		switch picture := elem.Photo.(type) {
		case *tg.PhotoEmpty: // photoEmpty#2331b22d
		case *tg.Photo: // photo#fb197a65
			inputPhotos = append(inputPhotos, picture.AsInput())
		}
	}
	_, err = client.API().PhotosDeletePhotos(ctx, inputPhotos)
	if err != nil {
		return err
	}
	return nil
}

// SetProfilePicture sets the current user profile picture.
// Use the `Uploader` struct in `github.com/gotd/td/telegram/uploader`
func SetProfilePicture(ctx context.Context, client *telegram.Client, file tg.InputFileClass) error {
	_, err := client.API().
		PhotosUploadProfilePhoto(ctx, &tg.PhotosUploadProfilePhotoRequest{
			File: file,
		})
	if err != nil {
		return err
	}
	return nil
}

// UpdateProfile updates common elements of the current user profile, like its username, first name,
// last name, and description.
func UpdateProfile(
	ctx context.Context,
	client *telegram.Client,
	username string,
	firstName string,
	lastName string,
	description string,
) error {
	_, err := client.API().AccountUpdateProfile(ctx, &tg.AccountUpdateProfileRequest{
		FirstName: firstName,
		LastName:  lastName,
		About:     description,
	})
	if err != nil {
		return err
	}
	_, err = client.API().AccountUpdateUsername(ctx, username)
	if err != nil {
		return err
	}
	return nil
}
