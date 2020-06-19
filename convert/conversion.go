// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package convert

// Conversion is an interface. A type implements Conversion will according
// the custom method to fill into database and retrieve from database.
type Conversion interface {
	To
	From
}

// From specifies the interface for scanning data from the database.
type From interface {
	FromDB([]byte) error
}

// To specifies the encoding data to sent to the database.
type To interface {
	ToDB() ([]byte, error)
}
