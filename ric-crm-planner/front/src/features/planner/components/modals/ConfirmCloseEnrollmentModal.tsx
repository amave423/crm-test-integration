import Modal from "../../../../components/Modal/Modal";
import AppButton from "../../../../components/UI/Button";

type ConfirmCloseEnrollmentModalProps = {
  isOpen: boolean;
  eventTitle?: string;
  onClose: () => void;
  onConfirm: () => void;
};

export default function ConfirmCloseEnrollmentModal({
  isOpen,
  eventTitle,
  onClose,
  onConfirm,
}: ConfirmCloseEnrollmentModalProps) {
  const resolvedTitle = eventTitle?.trim() || "это мероприятие";

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Подтверждение">
      <div className="confirm-body">
        <div className="confirm-text">
          Завершить набор по мероприятию «{resolvedTitle}»? Заявки со статусом «Добавился в орг. чат» перейдут в
          статус «Набор завершён», а CRM-робот отправит VK-сообщение с подтверждением участия.
        </div>
        <div className="confirm-actions">
          <AppButton className="link-btn" onClick={onClose}>
            Отмена
          </AppButton>
          <AppButton className="primary" onClick={onConfirm}>
            Подтвердить
          </AppButton>
        </div>
      </div>
    </Modal>
  );
}
