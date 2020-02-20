// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJoinLimit(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type Salary struct {
		Id  int64
		Lid int64
	}

	type CheckList struct {
		Id  int64
		Eid int64
	}

	type Empsetting struct {
		Id   int64
		Name string
	}

	assert.NoError(t, testEngine.Sync2(new(Salary), new(CheckList), new(Empsetting)))

	var emp Empsetting
	cnt, err := testEngine.Insert(&emp)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var checklist = CheckList{
		Eid: emp.Id,
	}
	cnt, err = testEngine.Insert(&checklist)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var salary = Salary{
		Lid: checklist.Id,
	}
	cnt, err = testEngine.Insert(&salary)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	tableName := tableMapper.Obj2Table("CheckList")
	tableName2 := tableMapper.Obj2Table("Salary")
	tableName3 := tableMapper.Obj2Table("Empsetting")

	idName := colMapper.Obj2Table("Id")
	lIDName := colMapper.Obj2Table("Lid")
	eIDName := colMapper.Obj2Table("Eid")

	var salaries []Salary
	err = testEngine.Table(tableName2).
		Join("INNER", tableName, tableName+"."+idName+" = "+tableName2+"."+lIDName).
		Join("LEFT", tableName3, tableName3+"."+idName+" = "+tableName+"."+eIDName).
		Limit(10, 0).
		Find(&salaries)
	assert.NoError(t, err)
}

func assertSync(t *testing.T, beans ...interface{}) {
	for _, bean := range beans {
		assert.NoError(t, testEngine.DropTables(bean))
		assert.NoError(t, testEngine.Sync2(bean))
	}
}

func TestWhere(t *testing.T) {
	assert.NoError(t, prepareEngine())

	assertSync(t, new(Userinfo))

	users := make([]Userinfo, 0)
	err := testEngine.Where("(id) > ?", 2).Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	fmt.Println(users)

	err = testEngine.Where("(id) > ?", 2).And("(id) < ?", 10).Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	fmt.Println(users)
}

func TestFind(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	users := make([]Userinfo, 0)

	err := testEngine.Find(&users)
	assert.NoError(t, err)
	for _, user := range users {
		fmt.Println(user)
	}

	users2 := make([]Userinfo, 0)
	var tbName = testEngine.Quote(testEngine.TableName(new(Userinfo), true))
	err = testEngine.SQL("select * from " + tbName).Find(&users2)
	assert.NoError(t, err)
}

func TestFind2(t *testing.T) {
	assert.NoError(t, prepareEngine())
	users := make([]*Userinfo, 0)

	assertSync(t, new(Userinfo))

	err := testEngine.Find(&users)
	assert.NoError(t, err)

	for _, user := range users {
		fmt.Println(user)
	}
}

type Team struct {
	Id int64
}

type TeamUser struct {
	OrgId  int64
	Uid    int64
	TeamId int64
}

func (TeamUser) TableName() string {
	return tableMapper.Obj2Table("TeamUser")
}

func TestFind3(t *testing.T) {
	var teamUser = new(TeamUser)
	assert.NoError(t, prepareEngine())
	err := testEngine.Sync2(new(Team), teamUser)
	assert.NoError(t, err)

	tableName := tableMapper.Obj2Table("TeamUser")
	teamTableName := tableMapper.Obj2Table("Team")
	idName := colMapper.Obj2Table("Id")
	orgIDName := colMapper.Obj2Table("OrgId")
	uidName := colMapper.Obj2Table("Uid")
	teamIDName := colMapper.Obj2Table("TeamId")

	var teams []Team
	err = testEngine.Cols("`"+teamTableName+"`."+idName).
		Where("`"+tableName+"`."+orgIDName+"=?", 1).
		And("`"+tableName+"`."+uidName+"=?", 2).
		Join("INNER", "`"+tableName+"`", "`"+tableName+"`."+teamIDName+"=`"+teamTableName+"`."+idName).
		Find(&teams)
	assert.NoError(t, err)

	teams = make([]Team, 0)
	err = testEngine.Cols("`"+teamTableName+"`."+idName).
		Where("`"+tableName+"`."+orgIDName+"=?", 1).
		And("`"+tableName+"`."+uidName+"=?", 2).
		Join("INNER", teamUser, "`"+tableName+"`."+teamIDName+"=`"+teamTableName+"`."+idName).
		Find(&teams)
	assert.NoError(t, err)

	teams = make([]Team, 0)
	err = testEngine.Cols("`"+teamTableName+"`."+idName).
		Where("`"+tableName+"`."+orgIDName+"=?", 1).
		And("`"+tableName+"`."+uidName+"=?", 2).
		Join("INNER", []interface{}{teamUser}, "`"+tableName+"`."+teamIDName+"=`"+teamTableName+"`."+idName).
		Find(&teams)
	assert.NoError(t, err)

	teams = make([]Team, 0)
	err = testEngine.Cols("`"+teamTableName+"`.id").
		Where("`tu`."+orgIDName+"=?", 1).
		And("`tu`."+uidName+"=?", 2).
		Join("INNER", []string{tableName, "tu"}, "`tu`."+teamIDName+"=`"+teamTableName+"`."+idName).
		Find(&teams)
	assert.NoError(t, err)

	teams = make([]Team, 0)
	err = testEngine.Cols("`"+teamTableName+"`."+idName).
		Where("`tu`."+orgIDName+"=?", 1).
		And("`tu`."+uidName+"=?", 2).
		Join("INNER", []interface{}{tableName, "tu"}, "`tu`."+teamIDName+"=`"+teamTableName+"`."+idName).
		Find(&teams)
	assert.NoError(t, err)

	teams = make([]Team, 0)
	err = testEngine.Cols("`"+teamTableName+"`."+idName).
		Where("`tu`."+orgIDName+"=?", 1).
		And("`tu`."+uidName+"=?", 2).
		Join("INNER", []interface{}{teamUser, "tu"}, "`tu`."+teamIDName+"=`"+teamTableName+"`."+idName).
		Find(&teams)
	assert.NoError(t, err)
}

func TestFindMap(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	users := make(map[int64]Userinfo)
	err := testEngine.Find(&users)
	assert.NoError(t, err)

	for _, user := range users {
		fmt.Println(user)
	}
}

func TestFindMap2(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	users := make(map[int64]*Userinfo)
	err := testEngine.Find(&users)
	assert.NoError(t, err)

	for id, user := range users {
		fmt.Println(id, user)
	}
}

func TestDistinct(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	_, err := testEngine.Insert(&Userinfo{
		Username: "lunny",
	})
	assert.NoError(t, err)

	users := make([]Userinfo, 0)
	departname := testEngine.GetTableMapper().Obj2Table("Departname")
	err = testEngine.Distinct(departname).Find(&users)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(users))

	fmt.Println(users)

	type Depart struct {
		Departname string
	}

	users2 := make([]Depart, 0)
	err = testEngine.Distinct(departname).Table(new(Userinfo)).Find(&users2)
	assert.NoError(t, err)
	if len(users2) != 1 {
		fmt.Println(len(users2))
		t.Error(err)
		panic(errors.New("should be one record"))
	}
	fmt.Println(users2)
}

func TestOrder(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	idName := colMapper.Obj2Table("Id")
	userName := colMapper.Obj2Table("Username")
	heightName := colMapper.Obj2Table("Height")

	users := make([]Userinfo, 0)
	err := testEngine.OrderBy(idName + " desc").Find(&users)
	assert.NoError(t, err)
	fmt.Println(users)

	users2 := make([]Userinfo, 0)
	err = testEngine.Asc(idName, userName).Desc(heightName).Find(&users2)
	assert.NoError(t, err)
	fmt.Println(users2)
}

func TestGroupBy(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	idName := colMapper.Obj2Table("Id")
	userName := colMapper.Obj2Table("Username")

	users := make([]Userinfo, 0)
	err := testEngine.GroupBy(idName + ", " + userName).Find(&users)
	assert.NoError(t, err)
}

func TestHaving(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	userName := colMapper.Obj2Table("Username")

	users := make([]Userinfo, 0)
	err := testEngine.GroupBy(userName).Having(userName + "='xlw'").Find(&users)
	assert.NoError(t, err)
	fmt.Println(users)

	users = make([]Userinfo, 0)
	err = testEngine.Cols("id, username").GroupBy("username").Having("username='xlw'").Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	fmt.Println(users)
}

func TestFindInts(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	userinfo := testEngine.GetTableMapper().Obj2Table("Userinfo")
	idName := colMapper.Obj2Table("Id")

	var idsInt64 []int64
	err := testEngine.Table(userinfo).Cols(idName).Desc(idName).Find(&idsInt64)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(idsInt64)

	var idsInt32 []int32
	err = testEngine.Table(userinfo).Cols(idName).Desc(idName).Find(&idsInt32)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(idsInt32)

	var idsInt []int
	err = testEngine.Table(userinfo).Cols(idName).Desc(idName).Find(&idsInt)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(idsInt)

	var idsUint []uint
	err = testEngine.Table(userinfo).Cols(idName).Desc(idName).Find(&idsUint)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(idsUint)

	type MyInt int
	var idsMyInt []MyInt
	err = testEngine.Table(userinfo).Cols(idName).Desc(idName).Find(&idsMyInt)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(idsMyInt)
}

func TestFindStrings(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))
	userinfo := testEngine.GetTableMapper().Obj2Table("Userinfo")
	username := testEngine.GetColumnMapper().Obj2Table("Username")
	idName := colMapper.Obj2Table("Id")

	var idsString []string
	err := testEngine.Table(userinfo).Cols(username).Desc(idName).Find(&idsString)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(idsString)
}

func TestFindMyString(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))
	userinfo := testEngine.GetTableMapper().Obj2Table("Userinfo")
	username := testEngine.GetColumnMapper().Obj2Table("Username")
	idName := colMapper.Obj2Table("Id")

	var idsMyString []MyString
	err := testEngine.Table(userinfo).Cols(username).Desc(idName).Find(&idsMyString)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(idsMyString)
}

func TestFindInterface(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	userinfo := testEngine.GetTableMapper().Obj2Table("Userinfo")
	username := testEngine.GetColumnMapper().Obj2Table("Username")
	idName := colMapper.Obj2Table("Id")

	var idsInterface []interface{}
	err := testEngine.Table(userinfo).Cols(username).Desc(idName).Find(&idsInterface)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(idsInterface)
}

func TestFindSliceBytes(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	userinfo := testEngine.GetTableMapper().Obj2Table("Userinfo")
	idName := colMapper.Obj2Table("Id")

	var ids [][][]byte
	err := testEngine.Table(userinfo).Desc(idName).Find(&ids)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range ids {
		fmt.Println(record)
	}
}

func TestFindSlicePtrString(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	userinfo := testEngine.GetTableMapper().Obj2Table("Userinfo")
	idName := colMapper.Obj2Table("Id")

	var ids [][]*string
	err := testEngine.Table(userinfo).Desc(idName).Find(&ids)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range ids {
		fmt.Println(record)
	}
}

func TestFindMapBytes(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	userinfo := testEngine.GetTableMapper().Obj2Table("Userinfo")
	idName := colMapper.Obj2Table("Id")

	var ids []map[string][]byte
	err := testEngine.Table(userinfo).Desc(idName).Find(&ids)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range ids {
		fmt.Println(record)
	}
}

func TestFindMapPtrString(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	userinfo := testEngine.GetTableMapper().Obj2Table("Userinfo")
	idName := colMapper.Obj2Table("Id")

	var ids []map[string]*string
	err := testEngine.Table(userinfo).Desc(idName).Find(&ids)
	assert.NoError(t, err)
	for _, record := range ids {
		fmt.Println(record)
	}
}

func TestFindBit(t *testing.T) {
	type FindBitStruct struct {
		Id  int64
		Msg bool `xorm:"bit"`
	}

	assert.NoError(t, prepareEngine())
	assertSync(t, new(FindBitStruct))

	cnt, err := testEngine.Insert([]FindBitStruct{
		{
			Msg: false,
		},
		{
			Msg: true,
		},
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 2, cnt)

	var results = make([]FindBitStruct, 0, 2)
	err = testEngine.Find(&results)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(results))
}

func TestFindMark(t *testing.T) {

	type Mark struct {
		Mark1 string `xorm:"VARCHAR(1)"`
		Mark2 string `xorm:"VARCHAR(1)"`
		MarkA string `xorm:"VARCHAR(1)"`
	}

	assert.NoError(t, prepareEngine())
	assertSync(t, new(Mark))

	cnt, err := testEngine.Insert([]Mark{
		{
			Mark1: "1",
			Mark2: "2",
			MarkA: "A",
		},
		{
			Mark1: "1",
			Mark2: "2",
			MarkA: "A",
		},
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 2, cnt)

	var results = make([]Mark, 0, 2)
	err = testEngine.Find(&results)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(results))
}

func TestFindAndCountOneFunc(t *testing.T) {
	type FindAndCountStruct struct {
		Id      int64
		Content string
		Msg     bool `xorm:"bit"`
	}

	assert.NoError(t, prepareEngine())
	assertSync(t, new(FindAndCountStruct))

	cnt, err := testEngine.Insert([]FindAndCountStruct{
		{
			Content: "111",
			Msg:     false,
		},
		{
			Content: "222",
			Msg:     true,
		},
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 2, cnt)

	var results = make([]FindAndCountStruct, 0, 2)
	cnt, err = testEngine.FindAndCount(&results)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(results))
	assert.EqualValues(t, 2, cnt)

	results = make([]FindAndCountStruct, 0, 1)
	cnt, err = testEngine.Where("msg = ?", true).FindAndCount(&results)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(results))
	assert.EqualValues(t, 1, cnt)

	results = make([]FindAndCountStruct, 0, 1)
	cnt, err = testEngine.Where("msg = ?", true).Limit(1).FindAndCount(&results)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(results))
	assert.EqualValues(t, 1, cnt)

	results = make([]FindAndCountStruct, 0, 1)
	cnt, err = testEngine.Where("msg = ?", true).Select("id, content, msg").
		Limit(1).FindAndCount(&results)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(results))
	assert.EqualValues(t, 1, cnt)

	results = make([]FindAndCountStruct, 0, 1)
	cnt, err = testEngine.Where("msg = ?", true).Desc("id").
		Limit(1).FindAndCount(&results)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(results))
	assert.EqualValues(t, 1, cnt)
}

type FindMapDevice struct {
	Deviceid string `xorm:"pk"`
	Status   int
}

func (device *FindMapDevice) TableName() string {
	return "devices"
}

func TestFindMapStringId(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(FindMapDevice))

	cnt, err := testEngine.Insert(&FindMapDevice{
		Deviceid: "1",
		Status:   1,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	deviceIDs := []string{"1"}

	deviceMaps := make(map[string]*FindMapDevice, len(deviceIDs))
	err = testEngine.
		Where("status = ?", 1).
		In("deviceid", deviceIDs).
		Find(&deviceMaps)
	assert.NoError(t, err)

	deviceMaps2 := make(map[string]FindMapDevice, len(deviceIDs))
	err = testEngine.
		Where("status = ?", 1).
		In("deviceid", deviceIDs).
		Find(&deviceMaps2)
	assert.NoError(t, err)

	devices := make([]*FindMapDevice, 0, len(deviceIDs))
	err = testEngine.Find(&devices)
	assert.NoError(t, err)

	devices2 := make([]FindMapDevice, 0, len(deviceIDs))
	err = testEngine.Find(&devices2)
	assert.NoError(t, err)

	var device FindMapDevice
	has, err := testEngine.Get(&device)
	assert.NoError(t, err)
	assert.True(t, has)

	has, err = testEngine.Exist(&FindMapDevice{})
	assert.NoError(t, err)
	assert.True(t, has)

	cnt, err = testEngine.Count(new(FindMapDevice))
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	cnt, err = testEngine.ID("1").Update(&FindMapDevice{
		Status: 2,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	sum, err := testEngine.SumInt(new(FindMapDevice), "status")
	assert.NoError(t, err)
	assert.EqualValues(t, 2, sum)

	cnt, err = testEngine.ID("1").Delete(new(FindMapDevice))
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)
}

func TestFindExtends(t *testing.T) {
	type FindExtendsB struct {
		ID int64
	}

	type FindExtendsA struct {
		FindExtendsB `xorm:"extends"`
	}

	assert.NoError(t, prepareEngine())
	assertSync(t, new(FindExtendsA))

	cnt, err := testEngine.Insert(&FindExtendsA{
		FindExtendsB: FindExtendsB{},
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	cnt, err = testEngine.Insert(&FindExtendsA{
		FindExtendsB: FindExtendsB{},
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var results []FindExtendsA
	err = testEngine.Find(&results)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(results))
}

func TestFindExtends3(t *testing.T) {
	type FindExtendsCC struct {
		ID   int64
		Name string
	}

	type FindExtendsBB struct {
		FindExtendsCC `xorm:"extends"`
	}

	type FindExtendsAA struct {
		FindExtendsBB `xorm:"extends"`
	}

	assert.NoError(t, prepareEngine())
	assertSync(t, new(FindExtendsAA))

	cnt, err := testEngine.Insert(&FindExtendsAA{
		FindExtendsBB: FindExtendsBB{
			FindExtendsCC: FindExtendsCC{
				Name: "cc1",
			},
		},
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	cnt, err = testEngine.Insert(&FindExtendsAA{
		FindExtendsBB: FindExtendsBB{
			FindExtendsCC: FindExtendsCC{
				Name: "cc2",
			},
		},
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var results []FindExtendsAA
	err = testEngine.Find(&results)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(results))
}

func TestFindCacheLimit(t *testing.T) {
	type InviteCode struct {
		ID      int64     `xorm:"pk autoincr 'id'"`
		Code    string    `xorm:"unique"`
		Created time.Time `xorm:"created"`
	}

	assert.NoError(t, prepareEngine())
	assertSync(t, new(InviteCode))

	cnt, err := testEngine.Insert(&InviteCode{
		Code: "123456",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	cnt, err = testEngine.Insert(&InviteCode{
		Code: "234567",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	for i := 0; i < 8; i++ {
		var beans []InviteCode
		err = testEngine.Limit(1, 0).Find(&beans)
		assert.NoError(t, err)
		assert.EqualValues(t, 1, len(beans))
	}

	for i := 0; i < 8; i++ {
		var beans2 []*InviteCode
		err = testEngine.Limit(1, 0).Find(&beans2)
		assert.NoError(t, err)
		assert.EqualValues(t, 1, len(beans2))
	}
}

func TestFindJoin(t *testing.T) {
	type SceneItem struct {
		Type     int
		DeviceId int64
	}

	type DeviceUserPrivrels struct {
		UserId   int64
		DeviceId int64
	}

	assert.NoError(t, prepareEngine())
	assertSync(t, new(SceneItem), new(DeviceUserPrivrels))

	tableName1 := tableMapper.Obj2Table("SceneItem")
	tableName2 := tableMapper.Obj2Table("DeviceUserPrivrels")

	deviceIDName := colMapper.Obj2Table("DeviceId")
	userIDName := colMapper.Obj2Table("UserId")
	typeName := colMapper.Obj2Table("Type")

	var scenes []SceneItem
	err := testEngine.Join("LEFT OUTER", tableName2, tableName1+"."+deviceIDName+"="+tableName2+"."+deviceIDName).
		Where(tableName1+"."+typeName+"=?", 3).Or(tableName2+"."+userIDName+"=?", 339).Find(&scenes)
	assert.NoError(t, err)

	scenes = make([]SceneItem, 0)
	err = testEngine.Join("LEFT OUTER", new(DeviceUserPrivrels), tableName1+"."+deviceIDName+"="+tableName2+"."+deviceIDName).
		Where(tableName1+"."+typeName+"=?", 3).Or(tableName2+"."+userIDName+"=?", 339).Find(&scenes)
	assert.NoError(t, err)
}
