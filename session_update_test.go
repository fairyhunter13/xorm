// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpdateMap(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UpdateTable struct {
		Id   int64
		Name string
		Age  int
	}

	assert.NoError(t, testEngine.Sync2(new(UpdateTable)))
	var tb = UpdateTable{
		Name: "test",
		Age:  35,
	}
	_, err := testEngine.Insert(&tb)
	assert.NoError(t, err)

	tableName := tableMapper.Obj2Table("UpdateTable")
	nameName := colMapper.Obj2Table("Name")
	ageName := colMapper.Obj2Table("age")
	idName := colMapper.Obj2Table("Id")

	cnt, err := testEngine.Table(tableName).Where("`"+idName+"` = ?", tb.Id).Update(map[string]interface{}{
		nameName: "test2",
		ageName:  36,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)
}

func TestUpdateLimit(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UpdateTable2 struct {
		Id   int64
		Name string
		Age  int
	}

	assert.NoError(t, testEngine.Sync2(new(UpdateTable2)))
	var tb = UpdateTable2{
		Name: "test1",
		Age:  35,
	}
	cnt, err := testEngine.Insert(&tb)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	tb.Name = "test2"
	tb.Id = 0
	cnt, err = testEngine.Insert(&tb)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	nameName := "`" + colMapper.Obj2Table("Name") + "`"

	cnt, err = testEngine.OrderBy(nameName + " desc").Limit(1).Update(&UpdateTable2{
		Age: 30,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var uts []UpdateTable2
	err = testEngine.Find(&uts)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(uts))
	assert.EqualValues(t, 35, uts[0].Age)
	assert.EqualValues(t, 30, uts[1].Age)
}

type ForUpdate struct {
	Id   int64 `xorm:"pk"`
	Name string
}

func setupForUpdate(engine EngineInterface) error {
	v := new(ForUpdate)
	err := testEngine.DropTables(v)
	if err != nil {
		return err
	}
	err = testEngine.CreateTables(v)
	if err != nil {
		return err
	}

	list := []ForUpdate{
		{1, "data1"},
		{2, "data2"},
		{3, "data3"},
	}

	for _, f := range list {
		_, err = testEngine.Insert(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestForUpdate(t *testing.T) {
	if *ignoreSelectUpdate {
		return
	}

	err := setupForUpdate(testEngine)
	if err != nil {
		t.Error(err)
		return
	}

	session1 := testEngine.NewSession()
	session2 := testEngine.NewSession()
	session3 := testEngine.NewSession()
	defer session1.Close()
	defer session2.Close()
	defer session3.Close()

	// start transaction
	err = session1.Begin()
	if err != nil {
		t.Error(err)
		return
	}

	// use lock
	fList := make([]ForUpdate, 0)
	session1.ForUpdate()
	session1.Where("(id) = ?", 1)
	err = session1.Find(&fList)
	switch {
	case err != nil:
		t.Error(err)
		return
	case len(fList) != 1:
		t.Errorf("find not returned single row")
		return
	case fList[0].Name != "data1":
		t.Errorf("for_update.name must be `data1`")
		return
	}

	// wait for lock
	wg := &sync.WaitGroup{}

	// lock is used
	wg.Add(1)
	go func() {
		f2 := new(ForUpdate)
		session2.Where("(id) = ?", 1).ForUpdate()
		has, err := session2.Get(f2) // wait release lock
		switch {
		case err != nil:
			t.Error(err)
		case !has:
			t.Errorf("cannot find target row. for_update.id = 1")
		case f2.Name != "updated by session1":
			t.Errorf("read lock failed")
		}
		wg.Done()
	}()

	// lock is NOT used
	wg.Add(1)
	go func() {
		f3 := new(ForUpdate)
		session3.Where("(id) = ?", 1)
		has, err := session3.Get(f3) // wait release lock
		switch {
		case err != nil:
			t.Error(err)
		case !has:
			t.Errorf("cannot find target row. for_update.id = 1")
		case f3.Name != "data1":
			t.Errorf("read lock failed")
		}
		wg.Done()
	}()

	// wait for go rountines
	time.Sleep(50 * time.Millisecond)

	f := new(ForUpdate)
	f.Name = "updated by session1"
	session1.Where("(id) = ?", 1)
	session1.Update(f)

	// release lock
	err = session1.Commit()
	if err != nil {
		t.Error(err)
		return
	}

	wg.Wait()
}

func TestWithIn(t *testing.T) {
	type temp3 struct {
		Id   int64  `xorm:"Id pk autoincr"`
		Name string `xorm:"Name"`
		Test bool   `xorm:"Test"`
	}

	assert.NoError(t, prepareEngine())
	assert.NoError(t, testEngine.Sync(new(temp3)))

	testEngine.Insert(&[]temp3{
		{
			Name: "user1",
		},
		{
			Name: "user1",
		},
		{
			Name: "user1",
		},
	})

	cnt, err := testEngine.In("Id", 1, 2, 3, 4).Update(&temp3{Name: "aa"}, &temp3{Name: "user1"})
	assert.NoError(t, err)
	assert.EqualValues(t, 3, cnt)
}

type Condi map[string]interface{}

type UpdateAllCols struct {
	Id     int64
	Bool   bool
	String string
	Ptr    *string
}

type UpdateMustCols struct {
	Id     int64
	Bool   bool
	String string
}

type UpdateIncr struct {
	Id   int64
	Cnt  int
	Name string
}

type Article struct {
	Id      int32  `xorm:"pk INT autoincr"`
	Name    string `xorm:"VARCHAR(45)"`
	Img     string `xorm:"VARCHAR(100)"`
	Aside   string `xorm:"VARCHAR(200)"`
	Desc    string `xorm:"VARCHAR(200)"`
	Content string `xorm:"TEXT"`
	Status  int8   `xorm:"TINYINT(4)"`
}

func TestUpdateMap2(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(UpdateMustCols))

	tableName := tableMapper.Obj2Table("UpdateMustCols")
	_, err := testEngine.Table(tableName).Where("`"+colMapper.Obj2Table("Id")+"` =?", 1).Update(map[string]interface{}{
		colMapper.Obj2Table("Bool"): true,
	})
	assert.NoError(t, err)
}

func TestUpdate1(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	_, err := testEngine.Insert(&Userinfo{
		Username: "user1",
	})

	var ori Userinfo
	has, err := testEngine.Get(&ori)
	assert.NoError(t, err)
	assert.True(t, has)

	// update by id
	user := Userinfo{Username: "xxx", Height: 1.2}
	cnt, err := testEngine.ID(ori.Uid).Update(&user)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	userName := "`" + colMapper.Obj2Table("Username") + "`"
	heightName := "`" + colMapper.Obj2Table("Height") + "`"
	departName := "`" + colMapper.Obj2Table("Departname") + "`"
	detailIDName := "`detail_id`"
	isMan := "`" + colMapper.Obj2Table("IsMan") + "`"
	createdName := "`" + colMapper.Obj2Table("Created") + "`"

	condi := Condi{userName: "zzz", departName: ""}
	cnt, err = testEngine.Table(&user).ID(ori.Uid).Update(&condi)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	cnt, err = testEngine.Update(&Userinfo{Username: "yyy"}, &user)
	assert.NoError(t, err)

	total, err := testEngine.Count(&user)
	assert.NoError(t, err)
	assert.EqualValues(t, cnt, total)

	// nullable update
	{
		user := &Userinfo{Username: "not null data", Height: 180.5}
		_, err := testEngine.Insert(user)
		assert.NoError(t, err)

		userID := user.Uid
		has, err := testEngine.ID(userID).
			And(userName+" = ?", user.Username).
			And(heightName+" = ?", user.Height).
			And(departName+" = ?", "").
			And(detailIDName+" = ?", 0).
			And(isMan+" = ?", 0).
			Get(&Userinfo{})
		assert.NoError(t, err)
		assert.True(t, has)

		updatedUser := &Userinfo{Username: "null data"}
		cnt, err = testEngine.ID(userID).
			Nullable(heightName, departName, isMan, createdName).
			Update(updatedUser)
		assert.NoError(t, err)
		assert.EqualValues(t, 1, cnt)

		has, err = testEngine.ID(userID).
			And(userName+" = ?", updatedUser.Username).
			And(heightName+" IS NULL").
			And(departName+" IS NULL").
			And(isMan+" IS NULL").
			And(createdName+" IS NULL").
			And(detailIDName+" = ?", 0).
			Get(&Userinfo{})
		assert.NoError(t, err)
		assert.True(t, has)

		cnt, err = testEngine.ID(userID).Delete(&Userinfo{})
		assert.NoError(t, err)
		assert.EqualValues(t, 1, cnt)
	}

	err = testEngine.StoreEngine("Innodb").Sync2(&Article{})
	assert.NoError(t, err)

	defer func() {
		err = testEngine.DropTables(&Article{})
		assert.NoError(t, err)
	}()

	a := &Article{0, "1", "2", "3", "4", "5", 2}
	cnt, err = testEngine.Insert(a)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)
	assert.True(t, a.Id > 0)

	cnt, err = testEngine.ID(a.Id).Update(&Article{Name: "6"})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var s = "test"

	col1 := &UpdateAllCols{Ptr: &s}
	err = testEngine.Sync(col1)
	assert.NoError(t, err)

	_, err = testEngine.Insert(col1)
	assert.NoError(t, err)

	col2 := &UpdateAllCols{col1.Id, true, "", nil}
	_, err = testEngine.ID(col2.Id).AllCols().Update(col2)
	assert.NoError(t, err)

	col3 := &UpdateAllCols{}
	has, err = testEngine.ID(col2.Id).Get(col3)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, *col2, *col3)

	{
		col1 := &UpdateMustCols{}
		err = testEngine.Sync(col1)
		assert.NoError(t, err)

		_, err = testEngine.Insert(col1)
		assert.NoError(t, err)

		col2 := &UpdateMustCols{col1.Id, true, ""}
		boolStr := testEngine.GetColumnMapper().Obj2Table("Bool")
		stringStr := testEngine.GetColumnMapper().Obj2Table("String")
		_, err = testEngine.ID(col2.Id).MustCols(boolStr, stringStr).Update(col2)
		assert.NoError(t, err)

		col3 := &UpdateMustCols{}
		has, err := testEngine.ID(col2.Id).Get(col3)
		assert.NoError(t, err)
		assert.True(t, has)
		assert.EqualValues(t, *col2, *col3)
	}
}

func TestUpdateIncrDecr(t *testing.T) {
	assert.NoError(t, prepareEngine())

	col1 := &UpdateIncr{
		Name: "test",
	}
	assert.NoError(t, testEngine.Sync(col1))

	_, err := testEngine.Insert(col1)
	assert.NoError(t, err)

	colName := testEngine.GetColumnMapper().Obj2Table("Cnt")

	cnt, err := testEngine.ID(col1.Id).Incr(colName).Update(col1)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	newCol := new(UpdateIncr)
	has, err := testEngine.ID(col1.Id).Get(newCol)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, 1, newCol.Cnt)

	cnt, err = testEngine.ID(col1.Id).Decr(colName).Update(col1)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	newCol = new(UpdateIncr)
	has, err = testEngine.ID(col1.Id).Get(newCol)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, 0, newCol.Cnt)

	cnt, err = testEngine.ID(col1.Id).Cols(colName).Incr(colName).Update(col1)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)
}

type UpdatedUpdate struct {
	Id      int64
	Updated time.Time `xorm:"updated"`
}

type UpdatedUpdate2 struct {
	Id      int64
	Updated int64 `xorm:"updated"`
}

type UpdatedUpdate3 struct {
	Id      int64
	Updated int `xorm:"updated bigint"`
}

type UpdatedUpdate4 struct {
	Id      int64
	Updated int `xorm:"updated"`
}

type UpdatedUpdate5 struct {
	Id      int64
	Updated time.Time `xorm:"updated bigint"`
}

func TestUpdateUpdated(t *testing.T) {
	assert.NoError(t, prepareEngine())

	di := new(UpdatedUpdate)
	err := testEngine.Sync2(di)
	if err != nil {
		t.Fatal(err)
	}

	_, err = testEngine.Insert(&UpdatedUpdate{})
	if err != nil {
		t.Fatal(err)
	}

	ci := &UpdatedUpdate{}
	_, err = testEngine.ID(1).Update(ci)
	if err != nil {
		t.Fatal(err)
	}

	has, err := testEngine.ID(1).Get(di)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(ErrNotExist)
	}
	if ci.Updated.Unix() != di.Updated.Unix() {
		t.Fatal("should equal:", ci, di)
	}
	fmt.Println("ci:", ci, "di:", di)

	di2 := new(UpdatedUpdate2)
	err = testEngine.Sync2(di2)
	assert.NoError(t, err)

	now := time.Now()
	var di20 UpdatedUpdate2
	cnt, err := testEngine.Insert(&di20)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)
	assert.True(t, now.Unix() <= di20.Updated)

	var di21 UpdatedUpdate2
	has, err = testEngine.ID(di20.Id).Get(&di21)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, di20.Updated, di21.Updated)

	ci2 := &UpdatedUpdate2{}
	_, err = testEngine.ID(1).Update(ci2)
	assert.NoError(t, err)

	has, err = testEngine.ID(1).Get(di2)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, ci2.Updated, di2.Updated)
	assert.True(t, ci2.Updated >= di21.Updated)

	di3 := new(UpdatedUpdate3)
	err = testEngine.Sync2(di3)
	if err != nil {
		t.Fatal(err)
	}

	_, err = testEngine.Insert(&UpdatedUpdate3{})
	if err != nil {
		t.Fatal(err)
	}
	ci3 := &UpdatedUpdate3{}
	_, err = testEngine.ID(1).Update(ci3)
	if err != nil {
		t.Fatal(err)
	}

	has, err = testEngine.ID(1).Get(di3)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(ErrNotExist)
	}
	if ci3.Updated != di3.Updated {
		t.Fatal("should equal:", ci3, di3)
	}
	fmt.Println("ci3:", ci3, "di3:", di3)

	di4 := new(UpdatedUpdate4)
	err = testEngine.Sync2(di4)
	if err != nil {
		t.Fatal(err)
	}

	_, err = testEngine.Insert(&UpdatedUpdate4{})
	if err != nil {
		t.Fatal(err)
	}

	ci4 := &UpdatedUpdate4{}
	_, err = testEngine.ID(1).Update(ci4)
	if err != nil {
		t.Fatal(err)
	}

	has, err = testEngine.ID(1).Get(di4)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(ErrNotExist)
	}
	if ci4.Updated != di4.Updated {
		t.Fatal("should equal:", ci4, di4)
	}
	fmt.Println("ci4:", ci4, "di4:", di4)

	di5 := new(UpdatedUpdate5)
	err = testEngine.Sync2(di5)
	if err != nil {
		t.Fatal(err)
	}

	_, err = testEngine.Insert(&UpdatedUpdate5{})
	if err != nil {
		t.Fatal(err)
	}
	ci5 := &UpdatedUpdate5{}
	_, err = testEngine.ID(1).Update(ci5)
	if err != nil {
		t.Fatal(err)
	}

	has, err = testEngine.ID(1).Get(di5)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(ErrNotExist)
	}
	if ci5.Updated.Unix() != di5.Updated.Unix() {
		t.Fatal("should equal:", ci5, di5)
	}
	fmt.Println("ci5:", ci5, "di5:", di5)
}

func TestUseBool(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	cnt1, err := testEngine.Count(&Userinfo{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	users := make([]Userinfo, 0)
	err = testEngine.Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	var fNumber int64
	for _, u := range users {
		if u.IsMan == false {
			fNumber += 1
		}
	}

	cnt2, err := testEngine.UseBool().Update(&Userinfo{IsMan: true})
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if fNumber != cnt2 {
		fmt.Println("cnt1", cnt1, "fNumber", fNumber, "cnt2", cnt2)
		/*err = errors.New("Updated number is not corrected.")
		  t.Error(err)
		  panic(err)*/
	}

	_, err = testEngine.Update(&Userinfo{IsMan: true})
	if err == nil {
		err = errors.New("error condition")
		t.Error(err)
		panic(err)
	}
}

func TestBool(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	_, err := testEngine.UseBool().Update(&Userinfo{IsMan: true})
	if err != nil {
		t.Error(err)
		panic(err)
	}
	users := make([]Userinfo, 0)
	err = testEngine.Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	for _, user := range users {
		if !user.IsMan {
			err = errors.New("update bool or find bool error")
			t.Error(err)
			panic(err)
		}
	}

	_, err = testEngine.UseBool().Update(&Userinfo{IsMan: false})
	if err != nil {
		t.Error(err)
		panic(err)
	}
	users = make([]Userinfo, 0)
	err = testEngine.Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	for _, user := range users {
		if user.IsMan {
			err = errors.New("update bool or find bool error")
			t.Error(err)
			panic(err)
		}
	}
}

func TestNoUpdate(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type NoUpdate struct {
		Id      int64
		Content string
	}

	assertSync(t, new(NoUpdate))

	cnt, err := testEngine.Insert(&NoUpdate{
		Content: "test",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	_, err = testEngine.ID(1).Update(&NoUpdate{})
	assert.Error(t, err)
	assert.EqualValues(t, "No content found to be updated", err.Error())
}

func TestNewUpdate(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type TbUserInfo struct {
		Id       int64       `xorm:"pk autoincr unique BIGINT" json:"id"`
		Phone    string      `xorm:"not null unique VARCHAR(20)" json:"phone"`
		UserName string      `xorm:"VARCHAR(20)" json:"user_name"`
		Gender   int         `xorm:"default 0 INTEGER" json:"gender"`
		Pw       string      `xorm:"VARCHAR(100)" json:"pw"`
		Token    string      `xorm:"TEXT" json:"token"`
		Avatar   string      `xorm:"TEXT" json:"avatar"`
		Extras   interface{} `xorm:"JSON" json:"extras"`
		Created  time.Time   `xorm:"DATETIME created"`
		Updated  time.Time   `xorm:"DATETIME updated"`
		Deleted  time.Time   `xorm:"DATETIME deleted"`
	}

	assertSync(t, new(TbUserInfo))

	targetUsr := TbUserInfo{Phone: "13126564922"}
	changeUsr := TbUserInfo{Token: "ABCDEFG"}
	af, err := testEngine.Update(&changeUsr, &targetUsr)
	assert.NoError(t, err)
	assert.EqualValues(t, 0, af)

	phoneName := "`" + colMapper.Obj2Table("Phone") + "`"

	af, err = testEngine.Table(new(TbUserInfo)).Where(phoneName+"=?", 13126564922).Update(&changeUsr)
	assert.NoError(t, err)
	assert.EqualValues(t, 0, af)
}

func TestUpdateUpdate(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type PublicKeyUpdate struct {
		Id          int64
		UpdatedUnix int64 `xorm:"updated"`
	}

	assertSync(t, new(PublicKeyUpdate))

	cnt, err := testEngine.ID(1).Cols("updated_unix").Update(&PublicKeyUpdate{
		UpdatedUnix: time.Now().Unix(),
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 0, cnt)
}

func TestCreatedUpdated2(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type CreatedUpdatedStruct struct {
		Id       int64
		Name     string
		CreateAt time.Time `xorm:"created" json:"create_at"`
		UpdateAt time.Time `xorm:"updated" json:"update_at"`
	}

	assertSync(t, new(CreatedUpdatedStruct))

	var s = CreatedUpdatedStruct{
		Name: "test",
	}
	cnt, err := testEngine.Insert(&s)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)
	assert.EqualValues(t, s.UpdateAt.Unix(), s.CreateAt.Unix())

	time.Sleep(time.Second)

	var s1 = CreatedUpdatedStruct{
		Name:     "test1",
		CreateAt: s.CreateAt,
		UpdateAt: s.UpdateAt,
	}

	cnt, err = testEngine.ID(1).Update(&s1)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)
	assert.EqualValues(t, s.CreateAt.Unix(), s1.CreateAt.Unix())
	assert.True(t, s1.UpdateAt.Unix() > s.UpdateAt.Unix())

	var s2 CreatedUpdatedStruct
	has, err := testEngine.ID(1).Get(&s2)
	assert.NoError(t, err)
	assert.True(t, has)

	assert.EqualValues(t, s.CreateAt.Unix(), s2.CreateAt.Unix())
	assert.True(t, s2.UpdateAt.Unix() > s.UpdateAt.Unix())
	assert.True(t, s2.UpdateAt.Unix() > s2.CreateAt.Unix())
}

func TestDeletedUpdate(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type DeletedUpdatedStruct struct {
		Id        int64
		Name      string
		DeletedAt time.Time `xorm:"deleted"`
	}

	assertSync(t, new(DeletedUpdatedStruct))

	var s = DeletedUpdatedStruct{
		Name: "test",
	}
	cnt, err := testEngine.Insert(&s)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	cnt, err = testEngine.ID(s.Id).Delete(&DeletedUpdatedStruct{})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	deletedAtName := colMapper.Obj2Table("DeletedAt")
	s.DeletedAt = time.Time{}
	cnt, err = testEngine.Unscoped().Nullable(deletedAtName).Update(&s)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var s1 DeletedUpdatedStruct
	has, err := testEngine.ID(s.Id).Get(&s1)
	assert.EqualValues(t, true, has)

	cnt, err = testEngine.ID(s.Id).Delete(&DeletedUpdatedStruct{})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	cnt, err = testEngine.ID(s.Id).Cols(deletedAtName).Update(&DeletedUpdatedStruct{})
	assert.EqualValues(t, "No content found to be updated", err.Error())
	assert.EqualValues(t, 0, cnt)

	cnt, err = testEngine.ID(s.Id).Unscoped().Cols(deletedAtName).Update(&DeletedUpdatedStruct{})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var s2 DeletedUpdatedStruct
	has, err = testEngine.ID(s.Id).Get(&s2)
	assert.EqualValues(t, true, has)
}

func TestUpdateMapCondition(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UpdateMapCondition struct {
		Id     int64
		String string
	}

	assertSync(t, new(UpdateMapCondition))

	var c = UpdateMapCondition{
		String: "string",
	}
	_, err := testEngine.Insert(&c)
	assert.NoError(t, err)

	idName := "`" + colMapper.Obj2Table("Id") + "`"

	cnt, err := testEngine.Update(&UpdateMapCondition{
		String: "string1",
	}, map[string]interface{}{
		idName: c.Id,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var c2 UpdateMapCondition
	has, err := testEngine.ID(c.Id).Get(&c2)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, "string1", c2.String)
}

func TestUpdateMapContent(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UpdateMapContent struct {
		Id     int64
		Name   string
		IsMan  bool
		Age    int
		Gender int // 1 is man, 2 is woman
	}

	assertSync(t, new(UpdateMapContent))

	var c = UpdateMapContent{
		Name:   "lunny",
		IsMan:  true,
		Gender: 1,
		Age:    18,
	}
	_, err := testEngine.Insert(&c)
	assert.NoError(t, err)
	assert.EqualValues(t, 18, c.Age)

	ageName := colMapper.Obj2Table("Age")
	isManName := colMapper.Obj2Table("IsMan")
	genderName := colMapper.Obj2Table("Gender")

	cnt, err := testEngine.Table(new(UpdateMapContent)).ID(c.Id).Update(map[string]interface{}{ageName: 0})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var c1 UpdateMapContent
	has, err := testEngine.ID(c.Id).Get(&c1)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, 0, c1.Age)

	cnt, err = testEngine.Table(new(UpdateMapContent)).ID(c.Id).Update(map[string]interface{}{
		ageName:    16,
		isManName:  false,
		genderName: 2,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var c2 UpdateMapContent
	has, err = testEngine.ID(c.Id).Get(&c2)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, 16, c2.Age)
	assert.EqualValues(t, false, c2.IsMan)
	assert.EqualValues(t, 2, c2.Gender)

	cnt, err = testEngine.Table(testEngine.TableName(new(UpdateMapContent))).ID(c.Id).Update(map[string]interface{}{
		ageName:    15,
		isManName:  true,
		genderName: 1,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var c3 UpdateMapContent
	has, err = testEngine.ID(c.Id).Get(&c3)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, 15, c3.Age)
	assert.EqualValues(t, true, c3.IsMan)
	assert.EqualValues(t, 1, c3.Gender)
}

func TestUpdateCondiBean(t *testing.T) {
	type NeedUpdateBean struct {
		Id   int64
		Name string
	}

	type NeedUpdateCondiBean struct {
		Name string
	}

	assert.NoError(t, prepareEngine())
	assertSync(t, new(NeedUpdateBean))

	cnt, err := testEngine.Insert(&NeedUpdateBean{
		Name: "name1",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	has, err := testEngine.Exist(&NeedUpdateBean{
		Name: "name1",
	})
	assert.NoError(t, err)
	assert.True(t, has)

	cnt, err = testEngine.Update(&NeedUpdateBean{
		Name: "name2",
	}, &NeedUpdateCondiBean{
		Name: "name1",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	has, err = testEngine.Exist(&NeedUpdateBean{
		Name: "name2",
	})
	assert.NoError(t, err)
	assert.True(t, has)

	cnt, err = testEngine.Update(&NeedUpdateBean{
		Name: "name1",
	}, NeedUpdateCondiBean{
		Name: "name2",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	has, err = testEngine.Exist(&NeedUpdateBean{
		Name: "name1",
	})
	assert.NoError(t, err)
	assert.True(t, has)
}

func TestWhereCondErrorWhenUpdate(t *testing.T) {
	type AuthRequestError struct {
		ChallengeToken string
		RequestToken   string
	}

	assert.NoError(t, prepareEngine())
	assertSync(t, new(AuthRequestError))

	_, err := testEngine.Cols("challenge_token", "request_token", "challenge_agent", "status").
		Where(&AuthRequestError{ChallengeToken: "1"}).
		Update(&AuthRequestError{
			ChallengeToken: "2",
		})
	assert.Error(t, err)
	assert.EqualValues(t, ErrConditionType, err)
}

func TestUpdateDeleted(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UpdateDeletedStruct struct {
		Id        int64
		Name      string
		DeletedAt time.Time `xorm:"deleted"`
	}

	assertSync(t, new(UpdateDeletedStruct))

	var s = UpdateDeletedStruct{
		Name: "test",
	}
	cnt, err := testEngine.Insert(&s)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	cnt, err = testEngine.ID(s.Id).Delete(&UpdateDeletedStruct{})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	cnt, err = testEngine.ID(s.Id).Update(&UpdateDeletedStruct{
		Name: "test1",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 0, cnt)

	nameName := colMapper.Obj2Table("Name")

	cnt, err = testEngine.Table(&UpdateDeletedStruct{}).ID(s.Id).Update(map[string]interface{}{
		nameName: "test1",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 0, cnt)

	cnt, err = testEngine.ID(s.Id).Unscoped().Update(&UpdateDeletedStruct{
		Name: "test1",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)
}

func TestUpdateExprs(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UpdateExprs struct {
		Id        int64
		NumIssues int
		Name      string
	}

	assertSync(t, new(UpdateExprs))

	_, err := testEngine.Insert(&UpdateExprs{
		NumIssues: 1,
		Name:      "lunny",
	})
	assert.NoError(t, err)

	_, err = testEngine.SetExpr("num_issues", "num_issues+1").AllCols().Update(&UpdateExprs{
		NumIssues: 3,
		Name:      "lunny xiao",
	})
	assert.NoError(t, err)

	var ue UpdateExprs
	has, err := testEngine.Get(&ue)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, 2, ue.NumIssues)
	assert.EqualValues(t, "lunny xiao", ue.Name)
}

func TestUpdateAlias(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UpdateAlias struct {
		Id        int64
		NumIssues int
		Name      string
	}

	assertSync(t, new(UpdateAlias))

	_, err := testEngine.Insert(&UpdateAlias{
		NumIssues: 1,
		Name:      "lunny",
	})
	assert.NoError(t, err)

	_, err = testEngine.Alias("ua").Where("ua.id = ?", 1).Update(&UpdateAlias{
		NumIssues: 2,
		Name:      "lunny xiao",
	})
	assert.NoError(t, err)

	var ue UpdateAlias
	has, err := testEngine.Get(&ue)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, 2, ue.NumIssues)
	assert.EqualValues(t, "lunny xiao", ue.Name)
}

func TestUpdateExprs2(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UpdateExprsRelease struct {
		Id         int64
		RepoId     int
		IsTag      bool
		IsDraft    bool
		NumCommits int
		Sha1       string
	}

	assertSync(t, new(UpdateExprsRelease))

	var uer = UpdateExprsRelease{
		RepoId:     1,
		IsTag:      false,
		IsDraft:    false,
		NumCommits: 1,
		Sha1:       "sha1",
	}
	inserted, err := testEngine.Insert(&uer)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, inserted)

	updated, err := testEngine.
		Where("repo_id = ? AND is_tag = ?", 1, false).
		SetExpr("is_draft", true).
		SetExpr("num_commits", 0).
		SetExpr("sha1", "").
		Update(new(UpdateExprsRelease))
	assert.NoError(t, err)
	assert.EqualValues(t, 1, updated)

	var uer2 UpdateExprsRelease
	has, err := testEngine.ID(uer.Id).Get(&uer2)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, 1, uer2.RepoId)
	assert.EqualValues(t, false, uer2.IsTag)
	assert.EqualValues(t, true, uer2.IsDraft)
	assert.EqualValues(t, 0, uer2.NumCommits)
	assert.EqualValues(t, "", uer2.Sha1)
}

func TestUpdateMap3(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UpdateMapUser struct {
		Id   uint64 `xorm:"PK autoincr"`
		Name string `xorm:""`
		Ver  uint64 `xorm:"version"`
	}

	oldMapper := testEngine.GetColumnMapper()
	defer func() {
		testEngine.SetColumnMapper(oldMapper)
	}()

	mapper := core.NewPrefixMapper(core.SnakeMapper{}, "F")
	testEngine.SetColumnMapper(mapper)

	assertSync(t, new(UpdateMapUser))

	_, err := testEngine.Table(new(UpdateMapUser)).Insert(map[string]interface{}{
		"Fname": "first user name",
		"Fver":  1,
	})
	assert.NoError(t, err)

	update := map[string]interface{}{
		"Fname": "user name",
		"Fver":  1,
	}
	rows, err := testEngine.Table(new(UpdateMapUser)).ID(1).Update(update)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, rows)

	update = map[string]interface{}{
		"Name": "user name",
		"Ver":  1,
	}
	rows, err = testEngine.Table(new(UpdateMapUser)).ID(1).Update(update)
	assert.Error(t, err)
	assert.EqualValues(t, 0, rows)
}
