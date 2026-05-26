import { Select as AntSelect } from "antd";
import type { SelectProps } from "antd";
import "../../styles/ui.scss";

type AppSelectTone = "event" | "directions" | "projects";
type AppSelectValue = string | number | Array<string | number>;

type AppSelectProps = SelectProps<AppSelectValue> & {
  tone?: AppSelectTone;
};

export default function Select({
  className = "",
  classNames,
  tone,
  popupMatchSelectWidth = false,
  listHeight = 260,
  showSearch = false,
  getPopupContainer,
  virtual,
  ...props
}: AppSelectProps) {
  const toneClass = tone ? `app-select--${tone}` : "";
  const popupToneClass = tone ? `app-select-dropdown--${tone}` : "";
  const popupRootClass = ["app-select-dropdown", popupToneClass, classNames?.popup?.root].filter(Boolean).join(" ");
  const isMobile = typeof window !== "undefined" && window.matchMedia("(max-width: 768px)").matches;
  const resolvedListHeight = isMobile ? Math.min(Number(listHeight) || 260, 190) : listHeight;
  const resolvedVirtual = typeof virtual === "boolean" ? virtual : !isMobile;
  const resolvePopupContainer =
    getPopupContainer ??
    ((triggerNode: HTMLElement) =>
      (triggerNode.closest(".modal") as HTMLElement | null) ??
      (triggerNode.closest(".wizard") as HTMLElement | null) ??
      (triggerNode.closest(".modal-content") as HTMLElement | null) ??
      document.body);

  return (
    <AntSelect
      {...props}
      popupMatchSelectWidth={popupMatchSelectWidth}
      listHeight={resolvedListHeight}
      showSearch={showSearch}
      getPopupContainer={resolvePopupContainer}
      virtual={resolvedVirtual}
      className={`app-select ${toneClass} ${className}`.trim()}
      classNames={{
        ...classNames,
        popup: {
          ...classNames?.popup,
          root: popupRootClass,
        },
      }}
    />
  );
}
