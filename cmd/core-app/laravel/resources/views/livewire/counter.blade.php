<div class="counter">
    <div class="counter__display">{{ $count }}</div>
    <div class="counter__controls">
        <button wire:click="decrement" class="btn btn--secondary">&minus;</button>
        <button wire:click="increment" class="btn btn--primary">+</button>
    </div>
    <p class="counter__hint">Livewire {{ \Livewire\Livewire::VERSION }} &middot; Server-rendered, no page reload</p>
</div>
