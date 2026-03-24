// output core functions

package iof

import "os"

// open menu from cli argument
func Go(menu string) {
  Cls()
  token, err := LoadToken()
  if True(err) || False(token) {
    GetLogin(SERVERENDPOINT)
  } else {
    ACCESSTOKEN = token
    me := ApiCall(Endpoint("me"), "")
    if False(me["found"]) {
      GetLogin(SERVERENDPOINT)
    } else {
      Setuser(String(me["name"]))
      Setzone(String(me["zone"]))
    }
  }
  Action(menu, nil)
}

// show user login info
func (c Standalone) Me() {
  route := Lower(Callup(0))
  Print(route, Endpoint(route), "")
}

// initialize client zone
// [An_] stands for anonymous
func (c Standalone) An_init_zone() {
  Zoner("init/zone")
}

// initialize client database
// [An_] stands for anonymous
func (c Standalone) An_init_accounting() {
  Zoner("init/accounting")
}

// launch direct api call
// [An_] stands for anonymous
func (c Standalone) An_direct_api_call() {
  Direct()
}

// post file data to server
// [An_] stands for anonymous
func (c Standalone) An_direct_api_post() {
  Post()
}

func (c Standalone) Login() {
  GetLogin(SERVERENDPOINT)
  os.Exit(0)
}

func (c Standalone) Logout() {
  if True(os.Remove(TOKENJSON)) {
    Error(ERRORDELETE())
  } else {
    Warning(TOKENJSON + " deleted")
  }
  os.Exit(0)
}

// show saved user info
func (c Standalone) Im() {
  token, _ := LoadToken()
  ACCESSTOKEN = token
  Print("me", Endpoint("me"), "")
  os.Exit(0)
}
