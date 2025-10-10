import { renderHook, act } from "@testing-library/react";
import { useMachineSelection } from "@/lib/useMachineSelection";

describe("useMachineSelection", () => {
    const STORAGE_KEY = "lunasentri_selected_machine_id";

    beforeEach(() => {
        // Clear localStorage before each test
        localStorage.clear();
        // Clear all mocks
        jest.clearAllMocks();
    });

    afterEach(() => {
        localStorage.clear();
    });

    describe("initialization", () => {
        it("should return null initially when no value is stored", () => {
            const { result } = renderHook(() => useMachineSelection());
            expect(result.current.selectedMachineId).toBeNull();
        });

        it("should load persisted value from localStorage on mount", () => {
            localStorage.setItem(STORAGE_KEY, "42");
            const { result } = renderHook(() => useMachineSelection());
            expect(result.current.selectedMachineId).toBe(42);
        });

        it("should handle invalid JSON in localStorage gracefully", () => {
            localStorage.setItem(STORAGE_KEY, "not-a-number");
            const { result } = renderHook(() => useMachineSelection());
            expect(result.current.selectedMachineId).toBeNull();
        });

        it("should handle non-numeric values in localStorage", () => {
            localStorage.setItem(STORAGE_KEY, JSON.stringify("abc"));
            const { result } = renderHook(() => useMachineSelection());
            expect(result.current.selectedMachineId).toBeNull();
        });
    });

    describe("selectMachine", () => {
        it("should update selectedMachineId and persist to localStorage", () => {
            const { result } = renderHook(() => useMachineSelection());

            act(() => {
                result.current.selectMachine(123);
            });

            expect(result.current.selectedMachineId).toBe(123);
            expect(localStorage.getItem(STORAGE_KEY)).toBe("123");
        });

        it("should update from one machine to another", () => {
            const { result } = renderHook(() => useMachineSelection());

            act(() => {
                result.current.selectMachine(10);
            });
            expect(result.current.selectedMachineId).toBe(10);

            act(() => {
                result.current.selectMachine(20);
            });
            expect(result.current.selectedMachineId).toBe(20);
            expect(localStorage.getItem(STORAGE_KEY)).toBe("20");
        });

        it("should handle selecting the same machine twice", () => {
            const { result } = renderHook(() => useMachineSelection());

            act(() => {
                result.current.selectMachine(5);
            });
            expect(result.current.selectedMachineId).toBe(5);

            act(() => {
                result.current.selectMachine(5);
            });
            expect(result.current.selectedMachineId).toBe(5);
            expect(localStorage.getItem(STORAGE_KEY)).toBe("5");
        });
    });

    describe("clearSelection", () => {
        it("should clear selection and remove from localStorage", () => {
            const { result } = renderHook(() => useMachineSelection());

            act(() => {
                result.current.selectMachine(99);
            });
            expect(result.current.selectedMachineId).toBe(99);

            act(() => {
                result.current.clearSelection();
            });

            expect(result.current.selectedMachineId).toBeNull();
            expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
        });

        it("should handle clearing when already null", () => {
            const { result } = renderHook(() => useMachineSelection());

            act(() => {
                result.current.clearSelection();
            });

            expect(result.current.selectedMachineId).toBeNull();
            expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
        });
    });

    describe("cross-tab synchronization", () => {
        it("should update state when localStorage changes in another tab", () => {
            const { result } = renderHook(() => useMachineSelection());

            expect(result.current.selectedMachineId).toBeNull();

            // Simulate storage event from another tab
            act(() => {
                localStorage.setItem(STORAGE_KEY, "777");
                window.dispatchEvent(
                    new StorageEvent("storage", {
                        key: STORAGE_KEY,
                        newValue: "777",
                        oldValue: null,
                        storageArea: localStorage,
                    })
                );
            });

            expect(result.current.selectedMachineId).toBe(777);
        });

        it("should handle clearing from another tab", () => {
            const { result } = renderHook(() => useMachineSelection());

            act(() => {
                result.current.selectMachine(50);
            });
            expect(result.current.selectedMachineId).toBe(50);

            // Simulate storage event clearing the value
            act(() => {
                localStorage.removeItem(STORAGE_KEY);
                window.dispatchEvent(
                    new StorageEvent("storage", {
                        key: STORAGE_KEY,
                        newValue: null,
                        oldValue: "50",
                        storageArea: localStorage,
                    })
                );
            });

            expect(result.current.selectedMachineId).toBeNull();
        });

        it("should ignore storage events for other keys", () => {
            const { result } = renderHook(() => useMachineSelection());

            act(() => {
                result.current.selectMachine(100);
            });

            // Simulate storage event for different key
            act(() => {
                window.dispatchEvent(
                    new StorageEvent("storage", {
                        key: "some_other_key",
                        newValue: "999",
                        oldValue: null,
                        storageArea: localStorage,
                    })
                );
            });

            // Should remain unchanged
            expect(result.current.selectedMachineId).toBe(100);
        });

        it("should handle invalid values in storage events", () => {
            const { result } = renderHook(() => useMachineSelection());

            act(() => {
                result.current.selectMachine(10);
            });

            // Simulate storage event with invalid value
            act(() => {
                window.dispatchEvent(
                    new StorageEvent("storage", {
                        key: STORAGE_KEY,
                        newValue: "not-a-number",
                        oldValue: "10",
                        storageArea: localStorage,
                    })
                );
            });

            // Should fall back to null for invalid values
            expect(result.current.selectedMachineId).toBeNull();
        });
    });

    describe("SSR safety", () => {
        it("should handle SSR environment gracefully", () => {
            const { result } = renderHook(() => useMachineSelection());

            // Hook should work without throwing errors
            expect(result.current.selectedMachineId).toBeNull();

            // Should handle selectMachine calls
            act(() => {
                result.current.selectMachine(123);
            });

            expect(result.current.selectedMachineId).toBe(123);
        });
    });

    describe("cleanup", () => {
        it("should remove storage event listener on unmount", () => {
            const removeEventListenerSpy = jest.spyOn(
                window,
                "removeEventListener"
            );
            const { unmount } = renderHook(() => useMachineSelection());

            unmount();

            expect(removeEventListenerSpy).toHaveBeenCalledWith(
                "storage",
                expect.any(Function)
            );
        });
    });
});
