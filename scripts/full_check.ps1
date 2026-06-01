$ErrorActionPreference = "Stop"
$base = "http://localhost:8080"
$results = New-Object System.Collections.Generic.List[object]

function Add-Result($name, $ok, $detail = "") {
  $script:results.Add([pscustomobject]@{ Name = $name; OK = $ok; Detail = "$detail" }) | Out-Null
}

function To-Json($obj) {
  $obj | ConvertTo-Json -Depth 20
}

function Api($name, $method, $url, $body = $null, $headers = @{}) {
  try {
    $params = @{ Uri = ($base + $url); Method = $method; Headers = $headers; TimeoutSec = 10 }
    if ($null -ne $body) {
      $params.ContentType = "application/json"
      $params.Body = To-Json $body
    }
    $res = Invoke-RestMethod @params
    $ok = ($null -ne $res.code -and [int]$res.code -eq 0) -or ($null -eq $res.code)
    Add-Result $name $ok ($(if ($res.msg) { $res.msg } else { "ok" }))
    return $res
  } catch {
    Add-Result $name $false $_.Exception.Message
    return $null
  }
}

function Find-One($items, $prop, $value) {
  $items | Where-Object { $_.$prop -eq $value } | Select-Object -First 1
}

$suffix = Get-Random -Minimum 10000 -Maximum 99999
$catName = "AUTO_TEST_CAT_$suffix"
$goodsName = "AUTO_TEST_GOODS_$suffix"
$storeName = "AUTO_TEST_STORE_$suffix"
$roleName = "AUTO_TEST_ROLE_$suffix"
$adminName = "AUTO_TEST_ADMIN_$suffix"
$adTitle = "AUTO_TEST_AD_$suffix"
$annTitle = "AUTO_TEST_ANN_$suffix"
$msgTitle = "AUTO_TEST_MSG_$suffix"

$cap = Api "admin captcha" "GET" "/admin/captcha?username=admin"
$login = Api "admin login" "POST" "/admin/login" @{ username = "admin"; password = "admin123"; captcha = $cap.data.captcha }
if (-not $login -or -not $login.data.token) {
  throw "admin login failed"
}
$adminHeaders = @{ Authorization = "Bearer $($login.data.token)" }

foreach ($p in @("/admin", "/admin/login", "/admin/commodity", "/admin/category", "/admin/order", "/admin/user", "/admin/admin-manager", "/admin/store", "/admin/role", "/admin/ads", "/admin/announcement", "/admin/message", "/mock-wx")) {
  try {
    $r = Invoke-WebRequest ($base + $p) -UseBasicParsing -TimeoutSec 10
    Add-Result "page $p" ($r.StatusCode -eq 200) $r.StatusCode
  } catch {
    Add-Result "page $p" $false $_.Exception.Message
  }
}

Api "admin dashboard" "GET" "/admin/dashboard" $null $adminHeaders | Out-Null

Api "category create" "POST" "/admin/categories" @{ name = $catName; sort = 88; icon = "" } $adminHeaders | Out-Null
$cats = Api "category list" "GET" "/admin/categories" $null $adminHeaders
$cat = Find-One $cats.data "name" $catName
Add-Result "category found after create" ($null -ne $cat) ($(if ($cat) { $cat.id } else { "not found" }))
$catId = [int]$cat.id
Api "category update" "PUT" "/admin/categories/$catId" @{ id = $catId; name = "${catName}_UPD"; sort = 89; icon = "" } $adminHeaders | Out-Null

Api "commodity create" "POST" "/admin/commodities" @{ name = $goodsName; category_id = $catId; price = 12.5; group_price = 10.5; stock = 30; image = "https://picsum.photos/seed/autotest/400/400"; description = "auto test goods"; status = 1; is_groupon = 1 } $adminHeaders | Out-Null
$coms = Api "commodity list" "GET" "/admin/commodities?page=1&limit=200" $null $adminHeaders
$com = Find-One $coms.data "name" $goodsName
Add-Result "commodity found after create" ($null -ne $com) ($(if ($com) { $com.id } else { "not found" }))
$comId = [int]$com.id
Api "commodity update" "PUT" "/admin/commodities/$comId" @{ id = $comId; name = "${goodsName}_UPD"; category_id = $catId; price = 13.5; group_price = 11.5; stock = 31; image = "https://picsum.photos/seed/autotest2/400/400"; description = "auto test goods updated"; status = 1; is_groupon = 1 } $adminHeaders | Out-Null
Api "commodity toggle" "PUT" "/admin/commodities/$comId/toggle" @{} $adminHeaders | Out-Null
Api "commodity toggle back" "PUT" "/admin/commodities/$comId/toggle" @{} $adminHeaders | Out-Null

Api "store create" "POST" "/admin/stores" @{ name = $storeName; address = "AUTO_TEST_ADDRESS"; phone = "13800000000"; leader_id = 0; status = 1 } $adminHeaders | Out-Null
$stores = Api "store list" "GET" "/admin/stores?page=1&limit=200" $null $adminHeaders
$store = Find-One $stores.data "name" $storeName
Add-Result "store found after create" ($null -ne $store) ($(if ($store) { $store.id } else { "not found" }))
$storeId = [int]$store.id
Api "store update" "PUT" "/admin/stores/$storeId" @{ id = $storeId; name = "${storeName}_UPD"; address = "AUTO_TEST_ADDRESS_UPD"; phone = "13900000000"; leader_id = 0; status = 0 } $adminHeaders | Out-Null
Api "store status back" "PUT" "/admin/stores/$storeId" @{ id = $storeId; name = "${storeName}_UPD"; address = "AUTO_TEST_ADDRESS_UPD"; phone = "13900000000"; leader_id = 0; status = 1 } $adminHeaders | Out-Null

Api "role create" "POST" "/admin/roles" @{ name = $roleName; desc = "AUTO_TEST_DESC"; menus = "1,2,3" } $adminHeaders | Out-Null
$roles = Api "role list" "GET" "/admin/roles" $null $adminHeaders
$role = Find-One $roles.data "name" $roleName
Add-Result "role found after create" ($null -ne $role) ($(if ($role) { $role.id } else { "not found" }))
$roleId = [int]$role.id
Api "role update" "PUT" "/admin/roles/$roleId" @{ id = $roleId; name = "${roleName}_UPD"; desc = "AUTO_TEST_DESC_UPD"; menus = "1,2" } $adminHeaders | Out-Null
Api "authorities list" "GET" "/admin/authorities?role_id=$roleId" $null $adminHeaders | Out-Null
Api "authorities save" "POST" "/admin/authorities" @{ role_id = $roleId; menu_ids = @(1, 2) } $adminHeaders | Out-Null

Api "admin create" "POST" "/admin/admins" @{ username = $adminName; password = "123456"; role_name = "AUTO_TEST_ROLE"; status = 1 } $adminHeaders | Out-Null
$admins = Api "admin list" "GET" "/admin/admins?page=1&limit=200" $null $adminHeaders
$newAdmin = Find-One $admins.data "username" $adminName
Add-Result "admin found after create" ($null -ne $newAdmin) ($(if ($newAdmin) { $newAdmin.id } else { "not found" }))
$newAdminId = [int]$newAdmin.id
Api "admin update" "PUT" "/admin/admins/$newAdminId" @{ role_name = "AUTO_TEST_ROLE_UPD"; status = 0 } $adminHeaders | Out-Null

Api "ads create" "POST" "/admin/ads" @{ title = $adTitle; image = "https://picsum.photos/seed/autoad/800/400"; link = "/mock-wx"; sort = 77; status = 1 } $adminHeaders | Out-Null
$ads = Api "ads list" "GET" "/admin/ads/all?page=1&limit=200" $null $adminHeaders
$ad = Find-One $ads.data "title" $adTitle
Add-Result "ads found after create" ($null -ne $ad) ($(if ($ad) { $ad.id } else { "not found" }))
$adId = [int]$ad.id
Api "ads update hide" "PUT" "/admin/ads/$adId" @{ id = $adId; title = "${adTitle}_UPD"; image = "https://picsum.photos/seed/autoad2/800/400"; link = "/mock-wx"; sort = 78; status = 0 } $adminHeaders | Out-Null

Api "announcement create" "POST" "/admin/announcements" @{ title = $annTitle; content = "AUTO_TEST_CONTENT"; status = 1 } $adminHeaders | Out-Null
$anns = Api "announcement list" "GET" "/admin/announcements/all?page=1&limit=200" $null $adminHeaders
$ann = Find-One $anns.data "title" $annTitle
Add-Result "announcement found after create" ($null -ne $ann) ($(if ($ann) { $ann.id } else { "not found" }))
$annId = [int]$ann.id
Api "announcement update hide" "PUT" "/admin/announcements/$annId" @{ id = $annId; title = "${annTitle}_UPD"; content = "AUTO_TEST_CONTENT_UPD"; status = 0 } $adminHeaders | Out-Null

$wxLogin = Api "wx login" "POST" "/wx/login" @{ code = "AUTO_TEST_WX_$suffix"; nickname = "AUTO_TEST_USER_$suffix"; avatar = "" } @{}
$wxHeaders = @{ Authorization = "Bearer $($wxLogin.data.token)" }
Api "wx profile get" "GET" "/wx/profile" $null $wxHeaders | Out-Null
Api "wx profile update" "PUT" "/wx/profile" @{ nickname = "AUTO_TEST_USER_${suffix}_UPD"; phone = "13600000000" } $wxHeaders | Out-Null
Api "wx categories" "GET" "/wx/categories" $null @{} | Out-Null
Api "wx commodities" "GET" "/wx/commodities" $null @{} | Out-Null
Api "wx commodity detail" "GET" "/wx/commodities/$comId" $null @{} | Out-Null
Api "wx ads public" "GET" "/wx/ads" $null @{} | Out-Null
Api "wx announcements public" "GET" "/wx/announcements" $null @{} | Out-Null
Api "wx stores public" "GET" "/wx/stores" $null @{} | Out-Null
Api "wx cart add" "POST" "/wx/cart" @{ commodity_id = $comId; quantity = 2 } $wxHeaders | Out-Null
$cart = Api "wx cart list" "GET" "/wx/cart" $null $wxHeaders
$item = $cart.data | Where-Object { $_.commodity_id -eq $comId } | Select-Object -First 1
Add-Result "cart item found after add" ($null -ne $item) ($(if ($item) { $item.id } else { "not found" }))
$itemId = [int]$item.id
Api "wx cart update qty" "PUT" "/wx/cart/$itemId" @{ quantity = 3; checked = 1 } $wxHeaders | Out-Null
Api "wx cart check all" "PUT" "/wx/cart/check-all" @{ checked = 1 } $wxHeaders | Out-Null
$order = Api "wx order create" "POST" "/wx/orders" @{ items = @($item); store_id = $storeId; remark = "AUTO_TEST_ORDER" } $wxHeaders
$orderId = [int]$order.data.id
Api "wx order list" "GET" "/wx/orders" $null $wxHeaders | Out-Null
Api "admin order list" "GET" "/admin/orders?page=1&limit=200" $null $adminHeaders | Out-Null
Api "admin order get" "GET" "/admin/orders/$orderId" $null $adminHeaders | Out-Null
Api "admin order items" "GET" "/admin/orders/$orderId/items" $null $adminHeaders | Out-Null
Api "admin order status pickup" "PUT" "/admin/orders/$orderId/status" @{ status = 2; store_id = $storeId } $adminHeaders | Out-Null
Api "wx order confirm" "PUT" "/wx/orders/$orderId/confirm" @{} $wxHeaders | Out-Null
Api "message create" "POST" "/admin/messages" @{ user_id = 0; title = $msgTitle; content = "AUTO_TEST_MESSAGE_CONTENT"; type = 0 } $adminHeaders | Out-Null
$msgs = Api "wx messages list" "GET" "/wx/messages" $null $wxHeaders
$msg = Find-One $msgs.data "title" $msgTitle
if ($msg) { Api "wx message read" "PUT" "/wx/messages/$([int]$msg.id)/read" @{} $wxHeaders | Out-Null } else { Add-Result "wx message found after send" $false "not found" }
Api "wx groupon subscribe" "POST" "/wx/groupon/subscribe" @{ commodity_id = $comId } $wxHeaders | Out-Null
$reminds = Api "wx groupon reminds list" "GET" "/wx/groupon/reminds" $null $wxHeaders
$rem = $reminds.data | Where-Object { $_.commodity_id -eq $comId } | Select-Object -First 1
if ($rem) { Api "wx groupon remind delete" "DELETE" "/wx/groupon/reminds/$([int]$rem.id)" $null $wxHeaders | Out-Null } else { Add-Result "groupon remind found after subscribe" $false "not found" }

Api "cleanup admin delete" "DELETE" "/admin/admins/$newAdminId" $null $adminHeaders | Out-Null
Api "cleanup ads delete" "DELETE" "/admin/ads/$adId" $null $adminHeaders | Out-Null
Api "cleanup announcement delete" "DELETE" "/admin/announcements/$annId" $null $adminHeaders | Out-Null
Api "cleanup role delete" "DELETE" "/admin/roles/$roleId" $null $adminHeaders | Out-Null
Api "cleanup commodity delete" "DELETE" "/admin/commodities/$comId" $null $adminHeaders | Out-Null
Api "cleanup category delete" "DELETE" "/admin/categories/$catId" $null $adminHeaders | Out-Null
Api "cleanup store delete" "DELETE" "/admin/stores/$storeId" $null $adminHeaders | Out-Null

$failed = $results | Where-Object { -not $_.OK }
[pscustomobject]@{
  Total = $results.Count
  Passed = ($results.Count - $failed.Count)
  Failed = $failed.Count
  Failures = $failed
} | ConvertTo-Json -Depth 10
