package admin_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor/admin"
	. "github.com/qor/admin/tests/dummy"
	qorTestUtils "github.com/qor/qor/test/utils"
	"github.com/theplant/htmltestingutils"
)

func TestEditPage(t *testing.T) {
	qorTestUtils.ResetDBTables(db, &admin.Group{}, &Language{}, &User{})
	user := createLoggedInUser()

	h := adminHandler

	var req *http.Request
	req, err := http.NewRequest("GET", fmt.Sprintf("/admin/users/%d/edit", user.ID), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	// assert only form so that other sections won't break the tests
	diff := htmltestingutils.PrettyHtmlDiff(rr.Body, "form.qor-form", expectedBody)
	if len(diff) > 0 {
		t.Error(diff)
	}
}

var expectedBody = `
<form class="qor-form" action="/admin/users/1" method="POST" enctype="multipart/form-data">
<input name="_method" value="PUT" type="hidden">

<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <input id="user_1_id" class="qor-hidden__primary_key" name="QorResource.ID" value="1" type="hidden">

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <div class="qor-field">
<div class="mdl-textfield mdl-textfield--full-width mdl-js-textfield">
<label class="qor-field__label mdl-textfield__label" for="user_1_name">
Name
</label>

<div class="qor-field__show">QOR</div>

<div class="qor-field__edit">
<input class="mdl-textfield__input" type="text" id="user_1_name" name="QorResource.Name" value="QOR" >
</div>
</div>
</div>

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <div class="qor-field">
<div class="mdl-textfield mdl-textfield--full-width mdl-js-textfield">
<label class="qor-field__label mdl-textfield__label" for="user_1_age">
Age
</label>

<div class="qor-field__show">
0
</div>

<div class="qor-field__edit">
<input class="mdl-textfield__input" type="number" id="user_1_age" name="QorResource.Age" value="0" >
</div>
</div>
</div>

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <div class="qor-field">
<div class="mdl-textfield mdl-textfield--full-width mdl-js-textfield">
<label class="qor-field__label mdl-textfield__label" for="user_1_role">
Role
</label>

<div class="qor-field__show">admin</div>

<div class="qor-field__edit">
<input class="mdl-textfield__input" type="text" id="user_1_role" name="QorResource.Role" value="admin" >
</div>
</div>
</div>

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <div class="qor-field">
<label class="mdl-checkbox mdl-js-checkbox mdl-js-ripple-effect" for="user_1_active">
<span class="qor-field__label mdl-checkbox__label">Active</span>

<span class="qor-field__edit">
<input type="checkbox" id="user_1_active" name="QorResource.Active" class="mdl-checkbox__input" value="true" type="checkbox" >
<input type="hidden" name="QorResource.Active" value="false">
</span>
</label>
</div>

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <div class="qor-field">
<div class="mdl-textfield mdl-textfield--full-width mdl-js-textfield">
<label class="qor-field__label mdl-textfield__label" for="user_1_registered_at">
Registered At
</label>

<div class="qor-field__show">

</div>

<div class="qor-field__edit qor-field__datetimepicker" data-picker-type="datetime">
<input class="mdl-textfield__input qor-datetimepicker__input" placeholder=" YYYY-MM-DD HH:MM " type="text" id="user_1_registered_at" name="QorResource.RegisteredAt" value="" >

<div>
	<button data-toggle="qor.datepicker" class="mdl-button mdl-js-button mdl-button--icon qor-action__datepicker" type="button">
	<i class="material-icons">date_range</i>
  </button>

  <button data-toggle="qor.timepicker" class="mdl-button mdl-js-button mdl-button--icon qor-action__timepicker" type="button">
	<i class="material-icons">access_time</i>
  </button>
</div>

</div>
</div>
</div>

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">


<div class="qor-field">
<label class="qor-field__label" for="user_1_avatar">
Avatar
</label>

<div class="qor-field__block qor-file ">
<div class="qor-fieldset">

<textarea class="qor-file__options hidden" data-cropper-title="Crop image" data-cropper-cancel="Cancel" data-cropper-ok="OK" name="QorResource.Avatar" aria-hidden="true">{&#34;FileName&#34;:&#34;&#34;,&#34;Url&#34;:&#34;&#34;}</textarea>
<div class="qor-file__list">

</div>

<label class="mdl-button mdl-button--primary qor-button__icon-add" title="Choose File" >
  <input class="visuallyhidden qor-file__input" id="user_1_avatar" name="QorResource.Avatar" type="file">
  Add Avatar
</label>

</div>
</div>
</div>

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">


<div class="signle-edit qor-field">
<label class="qor-field__label" for="user_1_profile">
Profile
</label>

<div class="qor-field__block">
<fieldset id="user_1_profile" class="qor-fieldset">
  <input id="" class="qor-hidden__primary_key" name="QorResource.Profile.ID" value="0" type="hidden">
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <input id="" class="qor-hidden__primary_key" name="QorResource.Profile.ID" value="0" type="hidden">

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <div class="qor-field">
<div class="mdl-textfield mdl-textfield--full-width mdl-js-textfield">
<label class="qor-field__label mdl-textfield__label" for="">
Name
</label>

<div class="qor-field__show"></div>

<div class="qor-field__edit">
<input class="mdl-textfield__input" type="text" id="" name="QorResource.Profile.Name" value="" >
</div>
</div>
</div>

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <div class="qor-field">
<div class="mdl-textfield mdl-textfield--full-width mdl-js-textfield">
<label class="qor-field__label mdl-textfield__label" for="">
Sex
</label>

<div class="qor-field__show"></div>

<div class="qor-field__edit">
<input class="mdl-textfield__input" type="text" id="" name="QorResource.Profile.Sex" value="" >
</div>
</div>
</div>

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">


<div class="signle-edit qor-field">
<label class="qor-field__label" for="">
Phone
</label>

<div class="qor-field__block">
<fieldset id="" class="qor-fieldset">
  <input id="" class="qor-hidden__primary_key" name="QorResource.Profile.Phone.ID" value="0" type="hidden">
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <input id="" class="qor-hidden__primary_key" name="QorResource.Profile.Phone.ID" value="0" type="hidden">

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <div class="qor-field">
<div class="mdl-textfield mdl-textfield--full-width mdl-js-textfield">
<label class="qor-field__label mdl-textfield__label" for="">
Num
</label>

<div class="qor-field__show"></div>

<div class="qor-field__edit">
<input class="mdl-textfield__input" type="text" id="" name="QorResource.Profile.Phone.Num" value="" >
</div>
</div>
</div>

</div>

</div>
</div>

</fieldset>
</div>
</div>


</div>

</div>
</div>

</fieldset>
</div>
</div>


</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">


<div class="signle-edit qor-field">
<label class="qor-field__label" for="user_1_credit_card">
Credit Card
</label>

<div class="qor-field__block">
<fieldset id="user_1_credit_card" class="qor-fieldset">
  <input id="" class="qor-hidden__primary_key" name="QorResource.CreditCard.ID" value="0" type="hidden">
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <input id="" class="qor-hidden__primary_key" name="QorResource.CreditCard.ID" value="0" type="hidden">

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <div class="qor-field">
<div class="mdl-textfield mdl-textfield--full-width mdl-js-textfield">
<label class="qor-field__label mdl-textfield__label" for="">
Number
</label>

<div class="qor-field__show"></div>

<div class="qor-field__edit">
<input class="mdl-textfield__input" type="text" id="" name="QorResource.CreditCard.Number" value="" >
</div>
</div>
</div>

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <div class="qor-field">
<div class="mdl-textfield mdl-textfield--full-width mdl-js-textfield">
<label class="qor-field__label mdl-textfield__label" for="">
Issuer
</label>

<div class="qor-field__show"></div>

<div class="qor-field__edit">
<input class="mdl-textfield__input" type="text" id="" name="QorResource.CreditCard.Issuer" value="" >
</div>
</div>
</div>

</div>

</div>
</div>

</fieldset>
</div>
</div>


</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">



<div class="qor-field collection-edit qor-fieldset-container" >
<label class="qor-field__label" for="user_1_addresses">
Addresses
</label>

<div class="qor-field__block">




<fieldset class="qor-fieldset qor-fieldset--new">
  <button data-confirm="Are you sure?" class="mdl-button qor-button--muted mdl-button--icon mdl-js-button qor-fieldset__delete" type="button">
	<i class="material-icons md-18">delete</i>
  </button>

	<input id="" class="qor-hidden__primary_key" name="QorResource.Addresses[0].ID" value="0" type="hidden">
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <input id="" class="qor-hidden__primary_key" name="QorResource.Addresses[0].ID" value="0" type="hidden">

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <div class="qor-field">
<div class="mdl-textfield mdl-textfield--full-width mdl-js-textfield">
<label class="qor-field__label mdl-textfield__label" for="">
Address1
</label>

<div class="qor-field__show"></div>

<div class="qor-field__edit">
<input class="mdl-textfield__input" type="text" id="" name="QorResource.Addresses[0].Address1" value="" >
</div>
</div>
</div>

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <div class="qor-field">
<div class="mdl-textfield mdl-textfield--full-width mdl-js-textfield">
<label class="qor-field__label mdl-textfield__label" for="">
Address2
</label>

<div class="qor-field__show"></div>

<div class="qor-field__edit">
<input class="mdl-textfield__input" type="text" id="" name="QorResource.Addresses[0].Address2" value="" >
</div>
</div>
</div>

</div>

</div>
</div>


</fieldset>

<button class="mdl-button mdl-button--primary qor-fieldset__add" type="button">
  Add Address
</button>

</div>
</div>

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">



<div class="qor-field">
<label class="qor-field__label" for="user_1_company">
Company

</label>

<div class="qor-field__show"></div>







<div class="qor-field__block qor-field__edit  qor-field__selectone" >

<select id="user_1_company" class="qor-field__input hidden"  data-toggle="qor.chooser" data-placeholder="Select an Option" name="QorResource.Company"   data-remote-url="/admin/companies" data-remote-data="true" data-remote-data-primary-key="ID">



</select>


<input type="hidden" name="QorResource.Company" value="">
</div>
</div>

</div>

</div>
</div>
<div class="qor-form-section clearfix" data-section-title="">


<div >

<div class="qor-form-section-rows qor-section-columns-1 clearfix">
  <div class="qor-field">
<label class="qor-field__label" for="user_1_languages">
Languages
</label>


<div class="qor-field__show qor-field__selectmany-show">

</div>








<div class="qor-field__edit qor-field__block qor-field__selectmany"  >

<select class="qor-field__input hidden" id="user_1_languages"  data-toggle="qor.chooser"  data-placeholder="Select some Options" name="QorResource.Languages" multiple  >



</select>


<input type="hidden" name="QorResource.Languages" value="">
</div>
</div>

</div>

</div>
</div>



  <div class="qor-form__actions">
	<button class="mdl-button mdl-button--colored mdl-button--raised mdl-js-button mdl-js-ripple-effect qor-button--save" type="submit">Save Changes</button>
	<a class="mdl-button mdl-button--primary mdl-js-button mdl-js-ripple-effect qor-button--cancel" href="javascript:history.back();">Cancel Edit</a>
  </div>

</form>
`
