param([string]$BaseUrl = 'http://127.0.0.1:8082')
$ErrorActionPreference = 'Stop'

function Request-Api {
    param([string]$Method, [string]$Path, [object]$Body = $null, [string]$Token = '', [int]$Expected = 200)
    $parameters = @{ Method = $Method; Uri = "$BaseUrl$Path"; UseBasicParsing = $true }
    if ($Token) { $parameters.Headers = @{ Authorization = "Bearer $Token" } }
    if ($null -ne $Body) {
        $parameters.ContentType = 'application/json; charset=utf-8'
        $parameters.Body = [System.Text.Encoding]::UTF8.GetBytes(($Body | ConvertTo-Json -Depth 10 -Compress))
    }
    try {
        $response = Invoke-WebRequest @parameters
        $status = [int]$response.StatusCode
    } catch {
        if (-not $_.Exception.Response) { throw }
        $status = [int]$_.Exception.Response.StatusCode
        if ($status -ne $Expected) { throw }
        Write-Host "$Method $Path -> $status"
        return
    }
    if ($status -ne $Expected) { throw "$Method $Path expected $Expected, received $status" }
    Write-Host "$Method $Path -> $status"
    if ($status -ne 204) { return ($response.Content | ConvertFrom-Json) }
}

$null = Request-Api GET '/healthz'
$suffix = [Guid]::NewGuid().ToString('N').Substring(0, 10)
$authorName = "smoke_$suffix"
$readerName = "reader_$suffix"
$author = Request-Api POST '/api/users' @{ user = @{ username = $authorName; email = "$authorName@example.com"; password = 'DemoPass123!' } } '' 201
$reader = Request-Api POST '/api/users' @{ user = @{ username = $readerName; email = "$readerName@example.com"; password = 'DemoPass123!' } } '' 201
$login = Request-Api POST '/api/users/login' @{ user = @{ email = "$authorName@example.com"; password = 'DemoPass123!' } }
$authorToken = $login.user.token
$readerToken = $reader.user.token
$created = Request-Api POST '/api/articles' @{ article = @{ title = 'HTTP smoke article'; description = 'Practical API verification'; body = 'A real request writes this article to SQLite.'; tagList = @('go', 'testing') } } $authorToken 201
$slug = $created.article.slug
$null = Request-Api PUT "/api/articles/$slug" @{ article = @{ title = 'Updated HTTP smoke article' } } $readerToken 403
$null = Request-Api PUT "/api/articles/$slug" @{ article = @{ title = 'Updated HTTP smoke article' } } $authorToken
$page = Request-Api GET "/api/articles?tag=go&author=$authorName&limit=1&offset=0"
if ($page.articlesCount -ne 1 -or $page.articles[0].slug -ne $slug) { throw 'Filter/page result mismatch' }
$comment = Request-Api POST "/api/articles/$slug/comments" @{ comment = @{ body = 'A useful real HTTP comment.' } } $readerToken 201
$commentId = $comment.comment.id
$null = Request-Api PUT "/api/articles/$slug/comments/$commentId" @{ comment = @{ body = 'Updated comment.' } } $readerToken
$comments = Request-Api GET "/api/articles/$slug/comments"
if ($comments.comments.Count -ne 1) { throw 'Comment missing' }
$null = Request-Api DELETE "/api/articles/$slug/comments/$commentId" $null $authorToken 403
1..2 | ForEach-Object {
    $favorite = Request-Api POST "/api/articles/$slug/favorite" $null $readerToken
    if ($favorite.article.favoritesCount -ne 1 -or -not $favorite.article.favorited) { throw 'Favorite was duplicated' }
}
$null = Request-Api DELETE "/api/articles/$slug/favorite" $null $readerToken
$null = Request-Api GET '/api/articles' $null 'broken-token' 401
$null = Request-Api DELETE "/api/articles/$slug/comments/$commentId" $null $readerToken 204
$null = Request-Api DELETE "/api/articles/$slug" $null $authorToken 204
$null = Request-Api GET "/api/articles/$slug" $null '' 404
Write-Host 'Smoke passed: auth, ownership, CRUD, filter/page, comment, unique favorite, cleanup.'
Write-Host 'Two example.com smoke users remain; generated article/comment/favorite are removed.'
