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

func (c UserCtx) EntriesFromCategory() {

}

func (c UserCtx) EntriesFromFeed() {

}

func (c UserCtx) EntriesFromTag() {

}

func (c UserCtx) Feeds() {

}

func (c UserCtx) FeedsFromCategory() {
}

func (c UserCtx) Categories() {

}

func (c UserCtx) Category() {

}

func (c UserCtx) Tags() {

}

func (c UserCtx) Tag() {

}
