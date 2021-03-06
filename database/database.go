/*
  Copyright (C) 2017 Jorge Martinez Hernandez

  This program is free software: you can redistribute it and/or modify
  it under the terms of the GNU Affero General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU Affero General Public License for more details.

  You should have received a copy of the GNU Affero General Public License
  along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

// Package database provides wrapper functions that create and modify objects based on models in an SQL database.
// See the models package for more information on the types of objects database operates on.
package database

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	mathRand "math/rand"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	// GORM dialect packages
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"golang.org/x/crypto/scrypt"

	"github.com/varddum/syndication/config"
	"github.com/varddum/syndication/models"
)

// Password salt and Hash byte sizes
const (
	PWSaltBytes = 32
	PWHashBytes = 64
)

// DB represents a connectin to a SQL database
type DB struct {
	db     *gorm.DB
	config config.Database
}

// DBError identifies error caused by database queries
type DBError interface {
	String() string
	Code() int
	Error() string
}

type (
	// Conflict is a DBError returned when a database operation
	// cannot be carried out because it conflicts with a previous operation.
	Conflict struct {
		msg string
	}

	// NotFound is DBError returned when an object cannot be found in the
	// database.
	NotFound struct {
		msg string
	}

	// BadRequest is a DBError returned when an operation is malformed.
	BadRequest struct {
		msg string
	}

	// Unauthorized is a DBError returned when a client does not have the permissions
	// to carry out an operation
	Unauthorized struct {
		msg string
	}

	// InternalError is a DBError returned when a client has failed to carry out
	// an operation and any other error type is not appropriate.
	InternalError struct {
		msg string
	}
)

// NewDB creates a new DB instance
func NewDB(conf config.Database) (db *DB, err error) {
	gormDB, err := gorm.Open(conf.Type, conf.Connection)
	if err != nil {
		return
	}

	db = &DB{
		config: conf,
	}

	gormDB.AutoMigrate(&models.Feed{})
	gormDB.AutoMigrate(&models.Category{})
	gormDB.AutoMigrate(&models.User{})
	gormDB.AutoMigrate(&models.Entry{})
	gormDB.AutoMigrate(&models.Tag{})
	gormDB.AutoMigrate(&models.APIKey{})

	db.db = gormDB

	return
}

var lastTimeIDWasCreated int64
var random32Int uint32

// Close ends connections with the database
func (db *DB) Close() error {
	return db.db.Close()
}

func createAPIID() string {
	currentTime := time.Now().Unix()
	duplicateTime := (lastTimeIDWasCreated == currentTime)
	lastTimeIDWasCreated = currentTime

	if !duplicateTime {
		random32Int = mathRand.Uint32() % 16
	} else {
		random32Int++
	}

	idStr := strconv.FormatInt(currentTime+int64(random32Int), 10)
	return base64.StdEncoding.EncodeToString([]byte(idStr))
}

func createPasswordHashAndSalt(password string) (hash []byte, salt []byte, err error) {
	salt = make([]byte, PWSaltBytes)
	_, err = io.ReadFull(rand.Reader, salt)
	if err != nil {
		return
	}

	hash, err = scrypt.Key([]byte(password), salt, 1<<14, 8, 1, PWHashBytes)
	if err != nil {
		return
	}

	return
}

// NewUser creates a new User object
func (db *DB) NewUser(username, password string) error {
	user := &models.User{}
	if !db.db.Where("username = ?", username).First(user).RecordNotFound() {
		return Conflict{"User already exists"}
	}

	hash, salt, err := createPasswordHashAndSalt(password)
	if err != nil {
		return err
	}

	// Construct the user system categories
	unctgAPIID := createAPIID()
	user.Categories = append(user.Categories, models.Category{
		APIID: unctgAPIID,
		Name:  models.Uncategorized,
	})
	user.UncategorizedCategoryAPIID = unctgAPIID

	user.APIID = createAPIID()
	user.PasswordHash = hash
	user.PasswordSalt = salt
	user.Username = username

	db.db.Create(&user).Related(&user.Categories)
	return nil
}

// DeleteUser deletes a User object
func (db *DB) DeleteUser(userID string) error {
	user := &models.User{}
	if db.db.Where("api_id = ?", userID).First(user).RecordNotFound() {
		return BadRequest{"User does not exists"}
	}

	db.db.Delete(user)
	return nil
}

// ChangeUserName for user with userID
func (db *DB) ChangeUserName(userID, newName string) error {
	user := &models.User{}
	if db.db.Where("api_id = ?", userID).First(user).RecordNotFound() {
		return BadRequest{"User does not exists"}
	}

	db.db.Model(user).Update("username", newName)
	return nil
}

// ChangeUserPassword for user with userID
func (db *DB) ChangeUserPassword(userID, newPassword string) error {
	user := &models.User{}
	if db.db.Where("api_id = ?", userID).First(user).RecordNotFound() {
		return BadRequest{"User does not exists"}
	}

	hash, salt, err := createPasswordHashAndSalt(newPassword)
	if err != nil {
		return err
	}

	db.db.Model(user).Update(models.User{
		PasswordHash: hash,
		PasswordSalt: salt,
	})
	return nil
}

// Users returns a list of all User entries.
// The parameter fields provides a way to select
// which fields are populated in the returned models.
func (db *DB) Users(fields ...string) (users []models.User) {
	selectFields := "id,api_id"
	if len(fields) != 0 {
		for _, field := range fields {
			selectFields = selectFields + "," + field
		}
	}
	db.db.Select(selectFields).Find(&users)
	return
}

// UserPrimaryKey returns the SQL primary key of a User with an api_id
func (db *DB) UserPrimaryKey(apiID string) (uint, error) {
	user := &models.User{}
	if db.db.First(user, "api_id = ?", apiID).RecordNotFound() {
		return 0, NotFound{"User does not exist"}
	}
	return user.ID, nil
}

// UserWithName returns a User with username
func (db *DB) UserWithName(username string) (user models.User, err error) {
	if db.db.First(&user, "username = ?", username).RecordNotFound() {
		err = NotFound{"User does not exist"}
	}
	return
}

// UserWithAPIID returns a User with id
func (db *DB) UserWithAPIID(apiID string) (user models.User, err error) {
	if db.db.First(&user, "api_id = ?", apiID).RecordNotFound() {
		err = NotFound{"User does not exist"}
	}
	return
}

// EntryWithAPIID returns an Entry with id that belongs to user
func (db *DB) EntryWithAPIID(apiID string, user *models.User) (entry models.Entry, err error) {
	if db.db.Model(user).First(&entry, "api_id = ?", apiID).RecordNotFound() {
		err = NotFound{"Entry does not exist"}
	}
	return
}

// Authenticate a user and return its respective User model if successful
func (db *DB) Authenticate(username, password string) (user models.User, err error) {
	user, err = db.UserWithName(username)
	if err != nil {
		err = Unauthorized{"Failed to authenticate user"}
		return
	}

	hash, err := scrypt.Key([]byte(password), user.PasswordSalt, 1<<14, 8, 1, PWHashBytes)
	if err != nil {
		err = InternalError{"Failed to authenticate user due to an internal error"}
		return
	}

	for i, hashByte := range hash {
		if hashByte != user.PasswordHash[i] {
			err = Unauthorized{"Failed to authenticate user"}
		}
	}

	return
}

// NewAPIKey creates a new APIKey object owned by user
func (db *DB) NewAPIKey(secret string, user *models.User) (models.APIKey, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.APIID
	claims["admin"] = false
	claims["exp"] = time.Now().Add(db.config.APIKeyExpiration.Duration).Unix()

	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return models.APIKey{}, err
	}

	key := &models.APIKey{
		Key:    t,
		User:   *user,
		UserID: user.ID,
	}

	db.db.Model(user).Association("APIKeys").Append(key)

	return *key, nil
}

// KeyBelongsToUser returns true if the given APIKey is owned by user
func (db *DB) KeyBelongsToUser(key *models.APIKey, user *models.User) (bool, error) {
	if key.Key == "" {
		return false, BadRequest{"No key provided"}
	}

	found := !db.db.Model(user).Where("key = ?", key.Key).Related(&models.APIKey{}).RecordNotFound()
	return found, nil
}

// NewFeed creates a new Feed object owned by user
func (db *DB) NewFeed(feed *models.Feed, user *models.User) error {
	feed.APIID = createAPIID()

	var err error
	var ctg models.Category
	if feed.Category.APIID != "" {
		ctg, err = db.Category(feed.Category.APIID, user)
		if err != nil {
			return BadRequest{"Feed has invalid category"}
		}
	} else {
		db.db.Model(user).Where("name = ?", models.Uncategorized).Related(&ctg)
	}

	feed.Category = ctg
	feed.CategoryID = ctg.ID
	feed.Category.APIID = ctg.APIID

	db.db.Model(user).Association("Feeds").Append(feed)
	db.db.Model(&ctg).Association("Feeds").Append(feed)

	return nil
}

// Feeds returns a list of all Feeds owned by a user
func (db *DB) Feeds(user *models.User) (feeds []models.Feed) {
	db.db.Model(user).Association("Feeds").Find(&feeds)
	return
}

// FeedsFromCategory returns all Feeds that belong to a category with categoryID
func (db *DB) FeedsFromCategory(categoryID string, user *models.User) (feeds []models.Feed, err error) {
	ctg, err := db.Category(categoryID, user)
	if err != nil {
		return
	}

	db.db.Model(ctg).Association("Feeds").Find(&feeds)
	return
}

// Feed returns a Feed with id and owned by user
func (db *DB) Feed(id string, user *models.User) (feed models.Feed, err error) {
	if db.db.Model(user).Where("api_id = ?", id).Related(&feed).RecordNotFound() {
		err = NotFound{"Feed does not exist"}
		return
	}

	db.db.Model(&feed).Related(&feed.Category)
	return
}

// DeleteFeed with id and owned by user
func (db *DB) DeleteFeed(id string, user *models.User) error {
	foundFeed := &models.Feed{}
	if !db.db.Model(user).Where("api_id = ?", id).Related(foundFeed).RecordNotFound() {
		db.db.Delete(foundFeed)
		return nil
	}
	return NotFound{"Feed does not exist"}
}

// EditFeed owned by user
func (db *DB) EditFeed(feed *models.Feed, user *models.User) error {
	foundFeed := &models.Feed{}
	if !db.db.Model(user).Where("api_id = ?", feed.APIID).Related(foundFeed).RecordNotFound() {
		foundFeed.Title = feed.Title
		db.db.Model(feed).Save(foundFeed)
		return nil
	}
	return NotFound{"Feed does not exist"}
}

// NewCategory creates a new Category object owned by user
func (db *DB) NewCategory(ctg *models.Category, user *models.User) error {
	if ctg.Name == "" {
		return BadRequest{"Category name should not be empty"}
	}

	tmpCtg := &models.Category{}
	if db.db.Model(user).Where("name = ?", ctg.Name).Related(tmpCtg).RecordNotFound() {
		ctg.APIID = createAPIID()
		db.db.Model(user).Association("Categories").Append(ctg)
		return nil
	}

	return Conflict{"Category already exists"}
}

// EditCategory owned by user
func (db *DB) EditCategory(ctg *models.Category, user *models.User) error {
	foundCtg := &models.Category{}
	if !db.db.Model(user).Where("api_id = ?", ctg.APIID).Related(foundCtg).RecordNotFound() {
		foundCtg.Name = ctg.Name
		db.db.Model(ctg).Save(foundCtg)
		return nil
	}
	return NotFound{"Category does not exist"}
}

// DeleteCategory with id and owned by user
func (db *DB) DeleteCategory(id string, user *models.User) error {
	if id == user.UncategorizedCategoryAPIID {
		return BadRequest{"Cannot delete system categories"}
	}

	ctg := &models.Category{}
	if db.db.Model(user).Where("api_id = ?", id).Related(ctg).RecordNotFound() {
		return NotFound{"Category does not exist"}
	}

	db.db.Delete(ctg)
	return nil
}

// Category returns a Category with id and owned by user
func (db *DB) Category(id string, user *models.User) (ctg models.Category, err error) {
	if db.db.Model(user).Where("api_id = ?", id).Related(&ctg).RecordNotFound() {
		err = NotFound{"Category does not exist"}
	}
	return
}

// Categories returns a list of all Categories owned by user
func (db *DB) Categories(user *models.User) (categories []models.Category) {
	db.db.Model(user).Association("Categories").Find(&categories)
	return
}

// ChangeFeedCategory changes the category a feed belongs to
func (db *DB) ChangeFeedCategory(feedID string, ctgID string, user *models.User) error {
	feed := &models.Feed{}
	if db.db.Model(user).Where("api_id = ?", feedID).Related(feed).RecordNotFound() {
		return NotFound{"Feed does not exist"}
	}

	prevCtg := &models.Category{
		ID: feed.CategoryID,
	}

	db.db.First(prevCtg)

	db.db.Model(prevCtg).Association("Feeds").Delete(feed)

	newCtg := &models.Category{}
	if db.db.Model(user).Where("api_id = ?", ctgID).Related(newCtg).RecordNotFound() {
		return NotFound{"Category does not exist"}
	}

	db.db.Model(newCtg).Association("Feeds").Append(feed)

	return nil
}

// NewEntry creates a new Entry object owned by user
func (db *DB) NewEntry(entry *models.Entry, user *models.User) error {
	if entry.Feed.APIID == "" {
		return BadRequest{"Entry should have a feed"}
	}

	feed := models.Feed{}
	if db.db.Model(user).Where("api_id = ?", entry.Feed.APIID).Related(&feed).RecordNotFound() {
		return NotFound{"Feed does not exist"}
	}

	entry.APIID = createAPIID()
	entry.Feed = feed
	entry.FeedID = feed.ID

	db.db.Model(user).Association("Entries").Append(entry)
	db.db.Model(&feed).Association("Entries").Append(entry)

	return nil
}

// NewEntries creates multiple new Entry objects which
// are all owned by feed with feedAPIID and user
func (db *DB) NewEntries(entries []models.Entry, feed *models.Feed, user *models.User) error {
	if feed.APIID == "" {
		return BadRequest{"Entry should have a feed"}
	}

	if len(entries) == 0 {
		return nil
	}

	if db.db.Model(user).Where("api_id = ?", feed.APIID).Related(feed).RecordNotFound() {
		return NotFound{"Feed does not exist"}
	}

	for _, entry := range entries {
		entry.APIID = createAPIID()

		db.db.Model(user).Association("Entries").Append(&entry)
		db.db.Model(feed).Association("Entries").Append(&entry)
	}

	return nil
}

// Entry returns an Entry with id and owned by user
func (db *DB) Entry(id string, user *models.User) (entry models.Entry, err error) {
	if db.db.Model(user).Where("api_id = ?", id).Related(&entry).RecordNotFound() {
		err = NotFound{"Feed does not exists"}
		return
	}

	db.db.Model(&entry).Related(&entry.Feed)
	return
}

// EntryWithGUIDExists returns true if an Entry exists with the given guid and is owned by user
func (db *DB) EntryWithGUIDExists(guid string, feedID string, user *models.User) (bool, error) {
	feed := &models.Feed{}
	if db.db.Model(user).Where("api_id = ?", feedID).Related(feed).RecordNotFound() {
		return true, NotFound{"Feed does not exist"}
	}

	return !db.db.Model(user).Where("guid = ? AND feed_id = ?", guid, feed.ID).Related(&models.Entry{}).RecordNotFound(), nil
}

// Entries returns a list of all entries owned by user
func (db *DB) Entries(orderByNewest bool, marker models.Marker, user *models.User) (entries []models.Entry, err error) {
	if marker == models.None {
		err = BadRequest{"Request should include a valid marker"}
		return
	}

	query := db.db.Model(user)
	if marker != models.Any {
		query = query.Where("mark = ?", marker)
	}

	if orderByNewest {
		query = query.Order("published DESC")
	} else {
		query = query.Order("published ASC")
	}

	query.Association("Entries").Find(&entries)
	return
}

// EntriesFromFeed returns all Entries that belong to a feed with feedID
func (db *DB) EntriesFromFeed(feedID string, orderByNewest bool, marker models.Marker, user *models.User) (entries []models.Entry, err error) {
	if marker == models.None {
		err = BadRequest{"Request should include a valid marker"}
		return
	}

	feed := &models.Feed{}
	if db.db.Model(user).Where("api_id = ?", feedID).Related(feed).RecordNotFound() {
		err = NotFound{"Feed not found"}
		return
	}

	query := db.db.Model(user)
	if marker != models.Any {
		query = query.Where("mark = ?", marker)
	}

	if orderByNewest {
		query = query.Order("published DESC")
	} else {
		query = query.Order("published ASC")
	}

	query.Where("feed_id = ?", feed.ID).Association("Entries").Find(&entries)

	return
}

// EntriesFromCategory returns all Entries that are related to a Category with categoryID by the entries' owning Feed
func (db *DB) EntriesFromCategory(categoryID string, orderByNewest bool, marker models.Marker, user *models.User) (entries []models.Entry, err error) {
	if marker == models.None {
		err = BadRequest{"Request should include a valid marker"}
		return
	}

	category := &models.Category{}
	if db.db.Model(user).Where("api_id = ?", categoryID).Related(category).RecordNotFound() {
		err = NotFound{"Category not found"}
		return
	}

	var feeds []models.Feed
	db.db.Model(category).Related(&feeds)

	query := db.db.Model(user)
	if marker != models.Any {
		query = query.Where("mark = ?", marker)
	}

	if orderByNewest {
		query = query.Order("published DESC")
	} else {
		query = query.Order("published ASC")
	}

	feedIds := make([]uint, len(feeds))
	for i, feed := range feeds {
		feedIds[i] = feed.ID
	}

	query.Where("feed_id in (?)", feedIds).Association("Entries").Find(&entries)
	return
}

// NewTag creates a new Tag object owned by user
func (db *DB) NewTag(tag *models.Tag, user *models.User) error {
	if tag.Name == "" {
		return BadRequest{"Tag name should not be empty"}
	}

	tmpTag := &models.Tag{}
	if db.db.Model(user).Where("name = ?", tag.Name).Related(tmpTag).RecordNotFound() {
		tag.APIID = createAPIID()
		db.db.Model(user).Association("Tags").Append(tag)
		return nil
	}

	return Conflict{"Tag already exists"}
}

// Tag returns a Tag with id and owned by user
func (db *DB) Tag(id string, user *models.User) (tag models.Tag, err error) {
	if db.db.Model(user).Where("api_id = ?", id).Related(&tag).RecordNotFound() {
		err = NotFound{"Tag does not exist"}
	}

	return
}

// Tags returns a list of all Tags owned by user
func (db *DB) Tags(user *models.User) (tags []models.Tag) {
	db.db.Model(user).Association("Tags").Find(&tags)
	return
}

// TagEntries with the given tag for user
func (db *DB) TagEntries(tagID string, entries []string, user *models.User) error {
	if len(entries) == 0 {
		return nil
	}

	tag := &models.Tag{}
	if db.db.Model(user).Where("api_id = ?", tagID).Related(tag).RecordNotFound() {
		return NotFound{"Tag does not exist"}
	}

	dbEntries := make([]models.Entry, len(entries))
	for i, entry := range entries {
		dbEntry, err := db.EntryWithAPIID(entry, user)
		if err != nil {
			return err
		}

		dbEntries[i] = dbEntry
	}

	for _, entry := range dbEntries {
		db.db.Model(tag).Association("Entries").Append(&entry)
	}

	return nil
}

// EntriesFromTag returns all Entries which are tagged with tagID
func (db *DB) EntriesFromTag(tagID string, marker models.Marker, orderByNewest bool, user *models.User) (entries []models.Entry, err error) {
	if marker == models.None {
		err = BadRequest{"Request should include a valid marker"}
		return
	}

	tag := &models.Tag{}
	if db.db.Model(user).Where("api_id = ?", tagID).Related(tag).RecordNotFound() {
		err = NotFound{"Tag not found"}
		return
	}

	query := db.db.Model(tag)
	if marker != models.Any {
		query = query.Where("mark = ?", marker)
	}

	if orderByNewest {
		query = query.Order("published DESC")
	} else {
		query = query.Order("published ASC")
	}

	query.Association("Entries").Find(&entries)

	return
}

// EntriesFromMultipleTags returns all Entries that are related to a Category with categoryID by the entries' owning Feed
func (db *DB) EntriesFromMultipleTags(tagIDs []string, orderByNewest bool, marker models.Marker, user *models.User) (entries []models.Entry, err error) {
	var order *gorm.DB
	if orderByNewest {
		order = db.db.Table("entries").Select("entries.title").Order("created_at DESC")
	} else {
		order = db.db.Table("entries").Select("entries.title").Order("created_at ASC")
	}

	if marker != models.Any {
		order = order.Where("mark = ?", marker)
	}

	var tagPrimaryKeys []uint
	for _, tag := range tagIDs {
		key, keyErr := db.TagPrimaryKey(tag)
		if keyErr != nil {
			err = keyErr
			return
		}

		tagPrimaryKeys = append(tagPrimaryKeys, key)
	}

	order.Joins("inner join entry_tags ON entry_tags.entry_id = entries.id").Where("entry_tags.tag_id in (?)", tagPrimaryKeys).Scan(&entries)
	return
}

// TagPrimaryKey returns the SQL primary key of a Tag with an api_id
func (db *DB) TagPrimaryKey(apiID string) (uint, error) {
	tag := &models.Tag{}
	if db.db.First(tag, "api_id = ?", apiID).RecordNotFound() {
		return 0, NotFound{"Tag does not exist"}
	}
	return tag.ID, nil
}

// EntryPrimaryKey returns the SQL primary key of an Entry with api_id
func (db *DB) EntryPrimaryKey(apiID string) (uint, error) {
	entry := &models.Entry{}
	if db.db.First(entry, "api_id = ?", apiID).RecordNotFound() {
		return 0, NotFound{"Entry does not exist"}
	}
	return entry.ID, nil
}

// EditTag owned by user
func (db *DB) EditTag(tag *models.Tag, user *models.User) error {
	foundTag := &models.Tag{}
	if !db.db.Model(user).Where("api_id = ?", tag.APIID).Related(foundTag).RecordNotFound() {
		foundTag.Name = tag.Name
		db.db.Model(tag).Save(foundTag)
		return nil
	}
	return NotFound{"Tag does not exist"}
}

// DeleteTag with id and owned by user
func (db *DB) DeleteTag(id string, user *models.User) error {
	tag := &models.Tag{}
	if db.db.Model(user).Where("api_id = ?", id).Related(tag).RecordNotFound() {
		return NotFound{"Tag does not exist"}
	}

	db.db.Delete(tag)
	return nil
}

// CategoryStats returns all Stats for a Category with the given id and that is owned by user
func (db *DB) CategoryStats(id string, user *models.User) (stats models.Stats, err error) {
	ctg := &models.Category{}
	if db.db.Model(user).Where("api_id = ?", id).Related(ctg).RecordNotFound() {
		err = NotFound{"Category not found"}
		return
	}

	var feeds []models.Feed
	db.db.Model(&ctg).Association("Feeds").Find(&feeds)

	feedIds := make([]uint, len(feeds))
	for i, feed := range feeds {
		feedIds[i] = feed.ID
	}

	query := db.db.Model(user).Where("feed_id in (?)", feedIds)

	stats.Unread = query.Where("mark = ?", models.Unread).Association("Entries").Count()
	stats.Read = query.Where("mark = ?", models.Read).Association("Entries").Count()
	stats.Saved = query.Where("saved = ?", true).Association("Entries").Count()
	stats.Total = query.Association("Entries").Count()
	return
}

// FeedStats returns all Stats for a Feed with the given id and that is owned by user
func (db *DB) FeedStats(id string, user *models.User) (stats models.Stats, err error) {
	feed := &models.Feed{}
	if db.db.Model(user).Where("api_id = ?", id).Related(feed).RecordNotFound() {
		err = NotFound{"Feed not found"}
		return
	}

	stats.Unread = db.db.Model(user).Where("feed_id = ? AND mark = ?", feed.ID, models.Unread).Association("Entries").Count()
	stats.Read = db.db.Model(user).Where("feed_id = ? AND mark = ?", feed.ID, models.Read).Association("Entries").Count()
	stats.Saved = db.db.Model(user).Where("feed_id = ? AND saved = ?", feed.ID, true).Association("Entries").Count()
	stats.Total = db.db.Model(user).Where("feed_id = ?", feed.ID).Association("Entries").Count()
	return
}

// Stats returns all Stats for the given user
func (db *DB) Stats(user *models.User) (stats models.Stats) {
	stats.Unread = db.db.Model(user).Where("mark = ?", models.Unread).Association("Entries").Count()
	stats.Read = db.db.Model(user).Where("mark = ?", models.Read).Association("Entries").Count()
	stats.Saved = db.db.Model(user).Where("saved = ?", true).Association("Entries").Count()
	stats.Total = db.db.Model(user).Association("Entries").Count()
	return
}

// MarkFeed applies marker to a Feed with id and owned by user
func (db *DB) MarkFeed(id string, marker models.Marker, user *models.User) error {
	feed, err := db.Feed(id, user)
	if err != nil {
		return err
	}

	db.db.Model(&models.Entry{}).Where("user_id = ? AND feed_id = ?", user.ID, feed.ID).Update(models.Entry{Mark: marker})
	return nil
}

// MarkCategory applies marker to a category with id and owned by user
func (db *DB) MarkCategory(id string, marker models.Marker, user *models.User) error {
	ctg, err := db.Category(id, user)
	if err != nil {
		return err
	}

	var feeds []models.Feed
	db.db.Model(&ctg).Association("Feeds").Find(&feeds)

	feedIds := make([]uint, len(feeds))
	for i, feed := range feeds {
		feedIds[i] = feed.ID
	}

	db.db.Model(&models.Entry{}).Where("user_id = ?", user.ID).Where("feed_id in (?)", feedIds).Update(models.Entry{Mark: marker})
	return nil
}

// MarkEntry applies marker to an entry with id and owned by user
func (db *DB) MarkEntry(id string, marker models.Marker, user *models.User) error {
	entry, err := db.Entry(id, user)
	if err != nil {
		return err
	}

	db.db.Model(&entry).Update(models.Entry{Mark: marker})
	return nil
}

// DeleteAll records in the database
func (db *DB) DeleteAll() {
	db.db.Delete(&models.Feed{})
	db.db.Delete(&models.Category{})
	db.db.Delete(&models.User{})
	db.db.Delete(&models.Entry{})
	db.db.Delete(&models.Tag{})
	db.db.Delete(&models.APIKey{})
}

func (e Conflict) Error() string {
	return e.msg
}

func (e Conflict) String() string {
	return "Conflict"
}

// Code returns Conflict's corresponding error code
func (e Conflict) Code() int {
	return 409
}

func (e NotFound) Error() string {
	return e.msg
}

func (e NotFound) String() string {
	return "Not Found"
}

// Code returns NotFound's corresponding error code
func (e NotFound) Code() int {
	return 404
}

func (e BadRequest) Error() string {
	return e.msg
}

func (e BadRequest) String() string {
	return "Bad Request"
}

// Code returns BadRequest's corresponding error code
func (e BadRequest) Code() int {
	return 400
}

func (e Unauthorized) Error() string {
	return e.msg
}

func (e Unauthorized) String() string {
	return "Unauthorized"
}

// Code returns Unauthorized's corresponding error code
func (e Unauthorized) Code() int {
	return 401
}

func (e InternalError) Error() string {
	return e.msg
}

// Code returns InternalError's corresponding error code
func (e InternalError) Code() int {
	return 500
}

func (e InternalError) String() string {
	return "Internal Error"
}
