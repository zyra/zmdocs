package templates

const BaseTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{ .Title }}</title>
    {{ if .Description }}<meta name="description" content="{{ .Description }}">{{ end }}
    <link href="https://unpkg.com/tailwindcss@^1.0/dist/tailwind.min.css" rel="stylesheet">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="theme-color" content="#fafafa">
</head>
<body>
<nav role="navigation">
    <div class="container flex items-center justify-between flex-wrap p-6 mx-auto">
        <div class="flex items-center flex-shrink-0 text-indigo-500 mr-6">
            <span class="font-semibold text-xl tracking-tight">
                {{ .SiteTitle }}
            </span>
        </div>
        <div class="block">
            <button id="sidenav-open-button" class="flex items-center px-3 py-2 text-indigo-400 hover:text-indigo-500">
                <svg class="fill-current h-5 w-5" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><title>Menu</title><path d="M0 3h20v2H0V3zm0 6h20v2H0V9zm0 6h20v2H0v-2z"/></svg>
            </button>
            <button id="sidenav-close-button" class="flex items-center px-3 py-2 text-indigo-400 hover:text-indigo-500 hidden">
                <svg class="fill-current w-5 h-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"><path d="M10 8.586L2.929 1.515 1.515 2.929 8.586 10l-7.071 7.071 1.414 1.414L10 11.414l7.071 7.071 1.414-1.414L11.414 10l7.071-7.071-1.414-1.414L10 8.586z"></path></svg>
            </button>
        </div>
    </div>
</nav>
<div class="container mx-auto">
    <div class="flex">
        <nav class="w-full lg:w-1/5 p-6 hidden lg:block" id="sidenav" role="navigation">
            <ul class="list-reset" role="none">
                {{ range $index, $element := .MenuItems }}
                    {{ if $element.Group }}
                        <li role="none" class="mb-6">
                            <span class="font-semibold text-gray-500">{{ $element.Title }}</span>
                            <ul class="list-reset ml-1" role="none">
                                {{ range $ci, $ce := $element.Items }}
                                    <li role="none">
                                        <a href="{{ $ce.Link }}" role="link" aria-label="{{ $ce.Title }}" class="block p-1 {{ if $ce.Active }}text-indigo-600{{ else }}text-gray-600 hover:text-gray-700{{ end }}">{{ $ce.Title }}</a>
                                    </li>
                                {{ end }}
                            </ul>
                        </li>
                    {{ else }}
                        <li role="none">
                            <a href="{{ $element.Link }}" role="link" aria-label="{{ $element.Title }}" class="block p-1 {{ if $element.Active }}text-indigo-600{{ else }}text-gray-600 hover:text-gray-700{{ end }}">{{ $element.Title }}</a>
                        </li>
                    {{ end }}
                {{ end }}
            </ul>
        </nav>
        <div class="w-full p-6" id="content-container">
            {{ .Content }}
        </div>
    </div>
</div>

<script>
    (function () {
        const openButton = document.getElementById('sidenav-open-button');
        const closeButton  = document.getElementById('sidenav-close-button');
        const sidenav = document.getElementById('sidenav');
        const container = document.getElementById('content-container');

        openButton.addEventListener('click', () => {
            sidenav.classList.replace('hidden', 'block');
            container.classList.add('hidden');
            openButton.classList.add('hidden');
            closeButton.classList.remove('hidden');
        });

        closeButton.addEventListener('click', () => {
            sidenav.classList.replace('block', 'hidden');
            container.classList.remove('hidden');
            openButton.classList.remove('hidden');
            closeButton.classList.add('hidden');
        });
    })();
</script>
</body>
</html>
`