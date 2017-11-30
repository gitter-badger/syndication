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
package plugins

import (
	"github.com/varddum/syndication/database"
	"github.com/varddum/syndication/models"
)

type (
	APICtx struct {
		User *UserCtx
	}

	UserCtx struct {
		db   *database.DB
		user *models.User
	}
)

func NewUserCtx(db *database.DB, user *models.User) UserCtx {
	return UserCtx{db, user}
}

func (c APICtx) HasUser() bool {
	return c.User != nil
}

func (c UserCtx) Entries(orderByNewest bool, marker models.Marker) ([]models.Entry, error) {
	return c.db.Entries(orderByNewest, marker, c.user)
}

func (c UserCtx) EntriesFromCategory(categoryID string, orderByNewest bool, marker models.Marker) ([]models.Entry, error) {
	return c.db.EntriesFromCategory(categoryID, orderByNewest, marker, c.user)
}

func (c UserCtx) EntriesFromFeed(feedID string, orderByNewest bool, marker models.Marker) ([]models.Entry, error) {
	return c.db.EntriesFromFeed(feedID, orderByNewest, marker, c.user)
}

func (c UserCtx) EntriesFromTag(tagID string, orderByNewest bool, marker models.Marker) ([]models.Entry, error) {
	return c.db.EntriesFromTag(tagID, marker, orderByNewest, c.user)
}

func (c UserCtx) EntriesFromMultipleTags(tagIDs []string, orderByNewest bool, marker models.Marker) ([]models.Entry, error) {
	return c.db.EntriesFromMultipleTags(tagIDs, orderByNewest, marker, c.user)
}

func (c UserCtx) Entry(id string) (models.Entry, error) {
	return c.db.Entry(id, c.user)
}

func (c UserCtx) Feeds() []models.Feed {
	return c.db.Feeds(c.user)
}

func (c UserCtx) FeedsFromCategory(categoryID string) ([]models.Feed, error) {
	return c.db.FeedsFromCategory(categoryID, c.user)
}

func (c UserCtx) Feed(id string) (models.Feed, error) {
	return c.db.Feed(id, c.user)
}

func (c UserCtx) DeleteFeed(id string) error {
	return c.db.DeleteFeed(id, c.user)
}

func (c UserCtx) EditFeed(feed *models.Feed) error {
	return c.db.EditFeed(feed, c.user)
}

func (c UserCtx) Categories() []models.Category {
	return c.db.Categories(c.user)
}

func (c UserCtx) Category(id string) (models.Category, error) {
	return c.db.Category(id, c.user)
}

func (c UserCtx) EditCategory(ctg *models.Category) error {
	return c.db.EditCategory(ctg, c.user)
}

func (c UserCtx) DeleteCategory(id string, user *models.User) error {
	return c.db.DeleteCategory(id, c.user)
}

func (c UserCtx) ChangeFeedCategory(feedID, ctgID string) error {
	return c.db.ChangeFeedCategory(feedID, ctgID, c.user)
}

func (c UserCtx) Tags() []models.Tag {
	return c.db.Tags(c.user)
}

func (c UserCtx) Tag(id string) (models.Tag, error) {
	return c.db.Tag(id, c.user)
}

func (c UserCtx) EditTag(tag *models.Tag) error {
	return c.db.EditTag(tag, c.user)
}

func (c UserCtx) DeleteTag(id string) error {
	return c.db.DeleteTag(id, c.user)
}

func (c UserCtx) TagEntries(tagID string, entries []string) error {
	return c.db.TagEntries(tagID, entries, c.user)
}

func (c UserCtx) CategoryStats(id string) (models.Stats, error) {
	return c.db.CategoryStats(id, c.user)
}

func (c UserCtx) FeedStats(id string) (models.Stats, error) {
	return c.db.FeedStats(id, c.user)
}

func (c UserCtx) Stats() models.Stats {
	return c.db.Stats(c.user)
}

func (c UserCtx) MarkFeed(id string, marker models.Marker) error {
	return c.db.MarkFeed(id, marker, c.user)
}

func (c UserCtx) MarkCategory(id string, marker models.Marker) error {
	return c.db.MarkCategory(id, marker, c.user)
}

func (c UserCtx) MarkEntry(id string, marker models.Marker) error {
	return c.db.MarkEntry(id, marker, c.user)
}
