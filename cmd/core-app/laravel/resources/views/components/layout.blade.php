<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Core App</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            background: #0d1117;
            color: #e6edf3;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            gap: 32px;
            padding: 32px;
        }
        .card {
            background: #161b22;
            border: 1px solid #30363d;
            border-radius: 12px;
            padding: 48px;
            text-align: center;
            max-width: 600px;
            width: 100%;
        }
        h1 { font-size: 32px; margin-bottom: 8px; }
        h2 { font-size: 20px; margin-bottom: 16px; color: #8b949e; font-weight: 400; }
        .accent { color: #39d0d8; }
        .subtitle { color: #8b949e; font-size: 16px; margin-bottom: 24px; }
        .info-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 12px;
            margin-top: 24px;
            text-align: left;
        }
        .info-item {
            background: #21262d;
            border: 1px solid #30363d;
            border-radius: 8px;
            padding: 12px 16px;
        }
        .info-item__label { font-size: 11px; color: #8b949e; text-transform: uppercase; letter-spacing: 0.5px; }
        .info-item__value { font-size: 14px; margin-top: 4px; font-family: monospace; }
        .badge {
            display: inline-block;
            background: #238636;
            color: #fff;
            border-radius: 12px;
            padding: 4px 12px;
            font-size: 12px;
            font-weight: 600;
            margin-top: 20px;
        }
        .counter { text-align: center; }
        .counter__display {
            font-size: 72px;
            font-weight: 700;
            font-variant-numeric: tabular-nums;
            color: #39d0d8;
            line-height: 1;
            margin-bottom: 24px;
        }
        .counter__controls {
            display: flex;
            gap: 16px;
            justify-content: center;
        }
        .counter__hint {
            margin-top: 16px;
            font-size: 12px;
            color: #8b949e;
        }
        .btn {
            appearance: none;
            border: 1px solid #30363d;
            border-radius: 8px;
            padding: 12px 32px;
            font-size: 20px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.15s;
        }
        .btn:active { transform: scale(0.96); }
        .btn--primary {
            background: #238636;
            color: #fff;
            border-color: #2ea043;
        }
        .btn--primary:hover { background: #2ea043; }
        .btn--secondary {
            background: #21262d;
            color: #e6edf3;
        }
        .btn--secondary:hover { background: #30363d; }
    </style>
    @livewireStyles
</head>
<body>
    {{ $slot }}
    @livewireScripts
</body>
</html>
