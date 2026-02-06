<x-layout>
    <div class="card">
        <h1><span class="accent">Core App</span></h1>
        <p class="subtitle">Laravel {{ app()->version() }} running inside a native desktop window</p>

        <div class="info-grid">
            <div class="info-item">
                <div class="info-item__label">PHP</div>
                <div class="info-item__value">{{ PHP_VERSION }}</div>
            </div>
            <div class="info-item">
                <div class="info-item__label">Thread Safety</div>
                <div class="info-item__value">{{ PHP_ZTS ? 'ZTS (Yes)' : 'NTS (No)' }}</div>
            </div>
            <div class="info-item">
                <div class="info-item__label">SAPI</div>
                <div class="info-item__value">{{ php_sapi_name() }}</div>
            </div>
            <div class="info-item">
                <div class="info-item__label">Platform</div>
                <div class="info-item__value">{{ PHP_OS }} {{ php_uname('m') }}</div>
            </div>
            <div class="info-item">
                <div class="info-item__label">Database</div>
                <div class="info-item__value">SQLite {{ \SQLite3::version()['versionString'] }}</div>
            </div>
            <div class="info-item">
                <div class="info-item__label">Mode</div>
                <div class="info-item__value">{{ env('FRANKENPHP_WORKER') ? 'Octane Worker' : 'Standard' }}</div>
            </div>
        </div>

        <div class="badge">Single Binary &middot; No Server &middot; No Config</div>
    </div>

    <div class="card">
        <h2>Livewire Reactivity Test</h2>
        <livewire:counter />
    </div>
</x-layout>
