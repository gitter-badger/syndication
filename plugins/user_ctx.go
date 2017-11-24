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
)

type (
	UserCtx struct {
		db *database.DB
	}
)

func NewUserCtx() UserCtx {
	return UserCtx{}
}

func (c UserCtx) Entries() {

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
