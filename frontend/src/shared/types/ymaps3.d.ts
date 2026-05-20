/// <reference types="@yandex/ymaps3-types" />

declare global {
  interface Window {
    ymaps3?: typeof ymaps3;
  }
}

export {};
